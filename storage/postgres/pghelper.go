package postgres

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cshekharsharma/photon/utils/types"
	"github.com/pkg/errors"
)

var (
	helperPostgresConnector PostgresDbConnectorInterface = &PostgresDbConnector{}
	getReadColumnsHook                                   = func(rows *sql.Rows) ([]string, error) {
		return rows.Columns()
	}
	scanReadRowHook = func(rows *sql.Rows, dest ...any) error {
		return rows.Scan(dest...)
	}
	hashFprintfHook = fmt.Fprintf
)

type sqlCloser interface {
	Close() error
}

func closeSQLCloser(closer sqlCloser) {
	if err := closer.Close(); err != nil {
		return
	}
}

// ReadQueryInput mirrors your mysql helper pattern.
type ReadQueryInput struct {
	Query             string
	Params            []any
	CapitaliseColumns bool
}

// ExecuteReadQuery runs a read-only SELECT query.
// It returns []map[string]any where each value is typed (int64, bool, time.Time, string, nil, etc).
//
// NOTE: We do NOT prepare implicitly here. Preparing without using the prepared statement
// is wasteful and can break assumptions under PgBouncer. If you want explicit prepare,
// do it at the call site (or add an optional flag and execute via stmt).
func ExecuteReadQuery(dbctx *DBContext, queryInput ReadQueryInput) ([]map[string]any, error) {
	return ExecuteReadQueryContext(context.Background(), dbctx, queryInput)
}

func ExecuteReadQueryContext(ctx context.Context, dbctx *DBContext, queryInput ReadQueryInput) ([]map[string]any, error) {
	if dbctx == nil {
		return nil, fmt.Errorf("nil DB context provided, cannot execute the query")
	}

	// TODO(otel): span "postgres.helper.read" (queryInput.Query sanitized), args count, cluster, etc.
	rows, err := dbctx.QueryContext(ctx, queryInput.Query, queryInput.Params...)
	if err != nil {
		return nil, err
	}
	if rows != nil {
		defer closeSQLCloser(rows)
	}

	cols, err := getReadColumnsHook(rows)
	if err != nil {
		return nil, err
	}

	out := make([]map[string]any, 0, 32)

	for rows.Next() {
		values := make([]any, len(cols))
		scanArgs := make([]any, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := scanReadRowHook(rows, scanArgs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any, len(cols))
		for i, colName := range cols {
			if queryInput.CapitaliseColumns {
				colName = types.UCFirst(colName)
			}
			rowMap[colName] = normalizePgValue(values[i])
		}

		out = append(out, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// normalizePgValue normalizes driver-returned values to friendlier Go types.
// - []byte (common under database/sql for text-like types) -> string
// - others left as-is
func normalizePgValue(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(x)
	default:
		return x
	}
}

// ExecuteWriteQuery executes an INSERT/UPDATE/DELETE query.
// Note: Postgres generally does NOT support LastInsertId().
// Prefer `RETURNING id` and QueryRow/Scan for inserts that need IDs.
func ExecuteWriteQuery(dbctx *DBContext, query string, params []any) (int64, int64, error) {
	return ExecuteWriteQueryContext(context.Background(), dbctx, query, params)
}

func ExecuteWriteQueryContext(ctx context.Context, dbctx *DBContext, query string, params []any) (int64, int64, error) {
	if dbctx == nil {
		return 0, 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}

	// TODO(otel): span "postgres.helper.write"
	result, err := dbctx.ExecContext(ctx, query, params...)
	if err != nil {
		return 0, 0, fmt.Errorf("error executing query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching impacted rows: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		lastInsertID = 0
	}

	return rowsAffected, lastInsertID, nil
}

// MultiInsertFromStructsArray performs a bulk insert using a slice of structs,
// generating Postgres placeholders ($1..$N).
func MultiInsertFromStructsArray[T any](dbctx *DBContext, tableName string, data []T) (int64, error) {
	return MultiInsertFromStructsArrayContext(context.Background(), dbctx, tableName, data)
}

func MultiInsertFromStructsArrayContext[T any](ctx context.Context, dbctx *DBContext, tableName string, data []T) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("input data array is empty")
	}
	if dbctx == nil {
		return 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}

	query, allValues, err := generateMultiInsertQueriesFromStructArray(tableName, data)
	if err != nil {
		return 0, errors.Wrap(err, "error in generating query from input data")
	}

	// TODO(otel): span "postgres.helper.multi_insert"
	result, err := dbctx.ExecContext(ctx, query, allValues...)
	if err != nil {
		return 0, fmt.Errorf("error executing insert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error fetching impacted rows: %w", err)
	}

	return rowsAffected, nil
}

func generateMultiInsertQueriesFromStructArray[T any](tableName string, data []T) (string, []any, error) {
	tbl, err := safeIdent(tableName)
	if err != nil {
		return "", nil, err
	}

	firstRow := reflect.ValueOf(data[0])
	if firstRow.Kind() == reflect.Pointer {
		firstRow = firstRow.Elem()
	}
	if firstRow.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("expected struct input, got %s", firstRow.Kind())
	}

	t := firstRow.Type()

	estimatedFieldCount := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if dbTag := field.Tag.Get("db"); dbTag != "" {
			tagParts := strings.Split(dbTag, ",")
			if tagParts[0] != "-" {
				estimatedFieldCount++
			}
		}
	}

	fieldsMap := make(map[string]struct{}, estimatedFieldCount)
	fields := make([]string, 0, estimatedFieldCount)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		dbTag := f.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		tagParts := strings.Split(dbTag, ",")
		name := tagParts[0]
		if name == "-" {
			continue
		}
		if _, ok := fieldsMap[name]; ok {
			continue
		}
		fieldsMap[name] = struct{}{}
		fields = append(fields, name)
	}

	quotedCols := make([]string, 0, len(fields))
	for _, c := range fields {
		qc, err := safeIdent(c)
		if err != nil {
			return "", nil, err
		}
		quotedCols = append(quotedCols, qc)
	}

	valueTuples := make([]string, 0, len(data))
	allValues := make([]any, 0, len(data)*estimatedFieldCount)
	argPos := 1

	for _, row := range data {
		sv := reflect.ValueOf(row)
		if sv.Kind() == reflect.Pointer {
			sv = sv.Elem()
		}
		if sv.Kind() != reflect.Struct {
			return "", nil, fmt.Errorf("expected struct input, got %s", sv.Kind())
		}

		fieldValues := make(map[string]any, len(fields))
		fieldIsDefault := make(map[string]bool, len(fields))

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			dbTag := f.Tag.Get("db")
			if dbTag == "" {
				continue
			}
			tagParts := strings.Split(dbTag, ",")
			name := tagParts[0]
			if name == "-" {
				continue
			}
			val := sv.Field(i).Interface()
			omitempty := slices.Contains(tagParts, "omitempty")
			useDefault := omitempty && types.IsEmpty(val)

			if !useDefault && slices.Contains(tagParts, "marshaljson") && sv.Field(i).Kind() == reflect.Struct {
				jv, err := json.Marshal(val)
				if err != nil {
					return "", nil, err
				}
				val = string(jv)
			}

			fieldValues[name] = val
			fieldIsDefault[name] = useDefault
		}

		placeholders := make([]string, 0, len(fields))
		for _, name := range fields {
			if fieldIsDefault[name] {
				placeholders = append(placeholders, "DEFAULT")
				continue
			}
			placeholders = append(placeholders, "$"+strconv.Itoa(argPos))
			argPos++
			allValues = append(allValues, fieldValues[name])
		}

		valueTuples = append(valueTuples, "("+strings.Join(placeholders, ", ")+")")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tbl, strings.Join(quotedCols, ", "), strings.Join(valueTuples, ", "))
	return query, allValues, nil
}

// InsertFromStruct inserts one record into a table based on struct db tags.
// Uses Postgres placeholders ($1..$N).
func InsertFromStruct(dbctx *DBContext, tableName string, data any) (int64, int64, error) {
	return InsertFromStructContext(context.Background(), dbctx, tableName, data)
}

func InsertFromStructContext(ctx context.Context, dbctx *DBContext, tableName string, data any) (int64, int64, error) {
	if dbctx == nil {
		return 0, 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}

	tbl, err := safeIdent(tableName)
	if err != nil {
		return 0, 0, err
	}

	sv := reflect.ValueOf(data)
	if sv.Kind() == reflect.Pointer {
		sv = sv.Elem()
	}
	if sv.Kind() != reflect.Struct {
		return 0, 0, fmt.Errorf("expected struct input, got %s", sv.Kind())
	}

	t := sv.Type()
	cols := make([]string, 0)
	vals := make([]any, 0)
	phs := make([]string, 0)
	argPos := 1

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		dbTag := f.Tag.Get("db")
		if dbTag == "" {
			continue
		}

		tagParts := strings.Split(dbTag, ",")
		col := tagParts[0]
		if col == "-" {
			continue
		}

		value := sv.Field(i).Interface()
		omitempty := slices.Contains(tagParts, "omitempty")
		if omitempty && types.IsEmpty(value) {
			continue
		}

		if slices.Contains(tagParts, "marshaljson") && sv.Field(i).Kind() == reflect.Struct {
			jv, err := json.Marshal(value)
			if err != nil {
				return 0, 0, err
			}
			value = string(jv)
		}

		qc, err := safeIdent(col)
		if err != nil {
			return 0, 0, err
		}

		cols = append(cols, qc)
		phs = append(phs, "$"+strconv.Itoa(argPos))
		vals = append(vals, value)
		argPos++
	}

	if len(cols) == 0 {
		return 0, 0, fmt.Errorf("no insertable fields found (check db tags / omitempty)")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tbl, strings.Join(cols, ", "), strings.Join(phs, ", "))

	// TODO(otel): span "postgres.helper.insert_struct"
	res, execErr := dbctx.ExecContext(ctx, query, vals...)
	if execErr != nil {
		return 0, 0, execErr
	}

	rowsAffected, _ := res.RowsAffected()
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		lastInsertID = 0
	}
	return rowsAffected, lastInsertID, nil
}

// InsertFromMap inserts a record using map[column]value.
// Uses Postgres placeholders ($1..$N).
func InsertFromMap(dbctx *DBContext, tableName string, data map[string]any) (int64, int64, error) {
	return InsertFromMapContext(context.Background(), dbctx, tableName, data)
}

func InsertFromMapContext(ctx context.Context, dbctx *DBContext, tableName string, data map[string]any) (int64, int64, error) {
	if dbctx == nil {
		return 0, 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty data map")
	}

	tbl, err := safeIdent(tableName)
	if err != nil {
		return 0, 0, err
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cols := make([]string, 0, len(keys))
	phs := make([]string, 0, len(keys))
	vals := make([]any, 0, len(keys))

	for i, col := range keys {
		qc, err := safeIdent(col)
		if err != nil {
			return 0, 0, err
		}
		cols = append(cols, qc)
		phs = append(phs, "$"+strconv.Itoa(i+1))
		vals = append(vals, data[col])
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tbl, strings.Join(cols, ", "), strings.Join(phs, ", "))

	// TODO(otel): span "postgres.helper.insert_map"
	res, err := dbctx.ExecContext(ctx, query, vals...)
	if err != nil {
		return 0, 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	lastInsertID, e := res.LastInsertId()
	if e != nil {
		lastInsertID = 0
	}
	return rowsAffected, lastInsertID, nil
}

// UpdateFromMap updates rows using map[column]value and a WHERE clause.
// whereClause may contain either:
// - Postgres placeholders ($N), OR
// - '?' placeholders (which will be converted to $N starting at the correct position)
func UpdateFromMap(dbctx *DBContext, tableName string, data map[string]any, whereClause string, params ...any) (int64, error) {
	return UpdateFromMapContext(context.Background(), dbctx, tableName, data, whereClause, params...)
}

func UpdateFromMapContext(ctx context.Context, dbctx *DBContext, tableName string, data map[string]any, whereClause string, params ...any) (int64, error) {
	if dbctx == nil {
		return 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("empty update map")
	}

	tbl, err := safeIdent(tableName)
	if err != nil {
		return 0, err
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	setParts := make([]string, 0, len(keys))
	values := make([]any, 0, len(keys)+len(params))
	argPos := 1

	for _, col := range keys {
		qc, err := safeIdent(col)
		if err != nil {
			return 0, err
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", qc, argPos))
		values = append(values, data[col])
		argPos++
	}

	convertedWhere, whereVals, err := convertWhereParamsToPg(whereClause, argPos, params...)
	if err != nil {
		return 0, err
	}
	values = append(values, whereVals...)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tbl, strings.Join(setParts, ", "), convertedWhere)

	// TODO(otel): span "postgres.helper.update_map"
	res, err := dbctx.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}

func DeleteByPrimaryKey(dbctx *DBContext, tableName, pkColumn string, pkValue any) (int64, error) {
	return DeleteByPrimaryKeyContext(context.Background(), dbctx, tableName, pkColumn, pkValue)
}

func DeleteByPrimaryKeyContext(ctx context.Context, dbctx *DBContext, tableName, pkColumn string, pkValue any) (int64, error) {
	if dbctx == nil {
		return 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}

	tbl, err := safeIdent(tableName)
	if err != nil {
		return 0, err
	}
	pk, err := safeIdent(pkColumn)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tbl, pk)

	// TODO(otel): span "postgres.helper.delete_pk"
	res, err := dbctx.ExecContext(ctx, query, pkValue)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}

func SoftDeleteByPrimaryKey(dbctx *DBContext, tableName, deleteCol, pkCol string, value any) (int64, error) {
	return SoftDeleteByPrimaryKeyContext(context.Background(), dbctx, tableName, deleteCol, pkCol, value)
}

func SoftDeleteByPrimaryKeyContext(ctx context.Context, dbctx *DBContext, tableName, deleteCol, pkCol string, value any) (int64, error) {
	if dbctx == nil {
		return 0, fmt.Errorf("nil DB context provided, cannot execute the query")
	}

	tbl, err := safeIdent(tableName)
	if err != nil {
		return 0, err
	}
	del, err := safeIdent(deleteCol)
	if err != nil {
		return 0, err
	}
	pk, err := safeIdent(pkCol)
	if err != nil {
		return 0, err
	}

	// Keep bool default; if your schema uses int flags, pass 1 instead of true.
	query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", tbl, del, pk)

	// TODO(otel): span "postgres.helper.soft_delete"
	res, err := dbctx.ExecContext(ctx, query, true, value)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}

// GetParameterizedInClause generates a named IN clause like ":id1,:id2,..."
// and returns a map of placeholders to values.
// This keeps your mysql helper API style.
func GetParameterizedInClause[T any](columnName string, columnValueArray []T) (string, map[string]any) {
	inClauseParams := make(map[string]any, len(columnValueArray))
	var b strings.Builder

	for i, v := range columnValueArray {
		key := fmt.Sprintf(":%s%d", columnName, i+1)
		b.WriteString(key)
		b.WriteString(",")
		inClauseParams[key] = v
	}

	inClauseString := strings.TrimSuffix(b.String(), ",")
	return inClauseString, inClauseParams
}

// ConvertQueryAndNamedParams converts ":name" tokens into Postgres positional params ($1,$2,...)
// and returns ordered param values.
//
// This implementation is robust against:
// - Postgres casts like "col::int" (won't treat ::int as a param)
// - Single-quoted strings: 'text :not_a_param'
// - Dollar-quoted strings: $$ ... :not_a_param ... $$
//
// Constraints (intentional for safety):
// - Named tokens must be [A-Za-z0-9_]+ and be prefixed by single ':'
// - The params maps must use keys like ":name" (same as your mysql helper)
func ConvertQueryAndNamedParams(query string, params ...map[string]any) (string, []any) {
	allParams := make(map[string]any)
	for _, mp := range params {
		for k, v := range mp {
			allParams[k] = v
		}
	}

	var out strings.Builder
	out.Grow(len(query) + 16)

	ordered := make([]any, 0, 8)
	seen := make(map[string]int, 16) // ":id" -> position

	argPos := 1

	// states
	inSingleQuote := false
	inDollarQuote := false
	dollarTag := "" // includes the $...$ tag (e.g. "$$", "$tag$")
	i := 0

	for i < len(query) {
		ch := query[i]

		// Handle dollar-quoted strings start/end
		if !inSingleQuote {
			if !inDollarQuote && ch == '$' {
				tag, ok := parseDollarTag(query[i:])
				if ok {
					inDollarQuote = true
					dollarTag = tag
					out.WriteString(tag)
					i += len(tag)
					continue
				}
			} else if inDollarQuote && ch == '$' && dollarTag != "" && strings.HasPrefix(query[i:], dollarTag) {
				// end
				inDollarQuote = false
				out.WriteString(dollarTag)
				i += len(dollarTag)
				dollarTag = ""
				continue
			}
		}

		// Handle single-quoted strings start/end (ignore escaped '' inside)
		if !inDollarQuote && ch == '\'' {
			out.WriteByte(ch)
			if inSingleQuote {
				// check escaped ''
				if i+1 < len(query) && query[i+1] == '\'' {
					// still in string
					out.WriteByte(query[i+1])
					i += 2
					continue
				}
				inSingleQuote = false
				i++
				continue
			}
			inSingleQuote = true
			i++
			continue
		}

		// If inside any quoted region, copy verbatim
		if inSingleQuote || inDollarQuote {
			out.WriteByte(ch)
			i++
			continue
		}

		// Named param detection
		if ch == ':' {
			// Skip Postgres cast "::"
			if i+1 < len(query) && query[i+1] == ':' {
				out.WriteString("::")
				i += 2
				continue
			}

			// parse token name
			j := i + 1
			for j < len(query) {
				c := query[j]
				if (c >= 'a' && c <= 'z') ||
					(c >= 'A' && c <= 'Z') ||
					(c >= '0' && c <= '9') ||
					c == '_' {
					j++
					continue
				}
				break
			}

			// If no name, treat ":" literally
			if j == i+1 {
				out.WriteByte(':')
				i++
				continue
			}

			token := query[i:j] // includes ':'
			if pos, ok := seen[token]; ok {
				out.WriteString("$" + strconv.Itoa(pos))
				i = j
				continue
			}

			seen[token] = argPos
			out.WriteString("$" + strconv.Itoa(argPos))
			ordered = append(ordered, allParams[token])
			argPos++
			i = j
			continue
		}

		out.WriteByte(ch)
		i++
	}

	return out.String(), ordered
}

// HashKey creates a deterministic hash for query+args map (same intent as your mysql helper).
func HashKey(query string, args map[string]any) (string, error) {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(query)
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(":")
		if _, err := hashFprintfHook(&b, "%v", args[k]); err != nil {
			return "", err
		}
		b.WriteString("|")
	}

	sum := sha1.Sum([]byte(b.String()))
	return base64.URLEncoding.EncodeToString(sum[:]), nil
}

// --- internals ---

// safeIdent validates and quotes identifiers (table/column).
// Identifiers cannot be bound as query params, so we must ensure they are trusted.
// This helper:
// - allows schema.table or table.column style identifiers
// - ensures only [a-zA-Z0-9_\.] characters are used
// - quotes each segment with double quotes (Postgres standard)
//
// NOTE: Use ONLY for identifiers, never for expressions.
func safeIdent(ident string) (string, error) {
	ident = strings.TrimSpace(ident)
	if ident == "" {
		return "", fmt.Errorf("empty identifier")
	}

	for _, r := range ident {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '.' {
			continue
		}
		return "", fmt.Errorf("invalid identifier %q", ident)
	}

	parts := strings.Split(ident, ".")
	for i, p := range parts {
		if p == "" {
			return "", fmt.Errorf("invalid identifier %q", ident)
		}
		parts[i] = `"` + p + `"`
	}
	return strings.Join(parts, "."), nil
}

// convertWhereParamsToPg converts a WHERE clause using '?' placeholders into Postgres style,
// starting at a given placeholder index.
//
// It validates that number of '?' matches len(params).
func convertWhereParamsToPg(where string, start int, params ...any) (string, []any, error) {
	if !strings.Contains(where, "?") {
		// assume caller already provided $ placeholders
		return where, params, nil
	}

	count := strings.Count(where, "?")
	if count != len(params) {
		return "", nil, fmt.Errorf("postgres: where placeholder mismatch: found %d '?' but got %d params", count, len(params))
	}

	out := where
	pos := start
	for strings.Contains(out, "?") {
		out = strings.Replace(out, "?", "$"+strconv.Itoa(pos), 1)
		pos++
	}
	return out, params, nil
}

// parseDollarTag detects $...$ tag at the beginning of s.
// Returns (tag, true) if found.
// Examples: "$$", "$tag$"
func parseDollarTag(s string) (string, bool) {
	if len(s) < 2 || s[0] != '$' {
		return "", false
	}
	// find next '$'
	j := 1
	for j < len(s) && s[j] != '$' {
		// tag chars: letters, digits, underscore
		c := s[j]
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' {
			j++
			continue
		}
		return "", false
	}
	if j < len(s) && s[j] == '$' {
		return s[:j+1], true
	}
	return "", false
}

// Optional small helper for callers.
func nowUTC() time.Time { return time.Now().UTC() }
