package mysql

import (
	"database/sql"
	"strconv"
	"time"
)

func CastSQLStringToString(value interface{}) string {
	if v, ok := value.(sql.NullString); ok {
		if v.Valid {
			return v.String
		}
		return ""
	}
	// Cache case
	if m, ok := value.(map[string]interface{}); ok {
		if valid, vok := m["Valid"].(bool); vok && valid {
			if str, sok := m["String"].(string); sok {
				return str
			}
		}
	}
	return ""
}

func CastSQLStringToInt(value interface{}) int64 {
	if v, ok := value.(sql.NullString); ok {
		if v.Valid {
			if intval, err := strconv.ParseInt(v.String, 10, 64); err == nil {
				return intval
			}
			return 0
		}
		return 0
	}
	// Cache case
	if m, ok := value.(map[string]interface{}); ok {
		if valid, vok := m["Valid"].(bool); vok && valid {
			if str, sok := m["String"].(string); sok {
				if intval, err := strconv.ParseInt(str, 10, 64); err == nil {
					return intval
				}
				return 0
			}
		}
	}
	return 0
}

func CastSQLStringToFloat(value interface{}) float64 {
	if v, ok := value.(sql.NullString); ok {
		if v.Valid {
			if floatval, err := strconv.ParseFloat(v.String, 64); err == nil {
				return floatval
			}
			return 0.0
		}
		return 0.0
	}
	// Cache case
	if m, ok := value.(map[string]interface{}); ok {
		if valid, vok := m["Valid"].(bool); vok && valid {
			if str, sok := m["String"].(string); sok {
				if floatval, err := strconv.ParseFloat(str, 64); err == nil {
					return floatval
				}
				return 0.0
			}
		}
	}
	return 0.0
}

func CastSQLStringToTime(value interface{}) time.Time {
	defaultVal, _ := time.Parse(time.DateTime, "")

	if v, ok := value.(sql.NullString); ok {
		if !v.Valid {
			return defaultVal
		}

		if timeval, err := time.Parse(time.DateTime, v.String); err == nil {
			return timeval
		}

		return defaultVal
	}

	// Cache case
	if m, ok := value.(map[string]interface{}); ok {
		if valid, vok := m["Valid"].(bool); vok && valid {
			if str, sok := m["String"].(string); sok {
				if timeval, err := time.Parse(time.DateTime, str); err == nil {
					return timeval
				}
				return defaultVal
			}
		}
	}
	return defaultVal
}
