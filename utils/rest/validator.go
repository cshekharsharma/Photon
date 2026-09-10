package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// Validator holds a validator.Validate pointer that can be used to validate data according to struct tags.
type Validator struct {
	validate *validator.Validate
}

// New creates and returns a new Validator with a fresh instance of a validator.Validate.
func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateJson decodes a JSON payload from an HTTP request and validates its structure.
// The function expects the JSON to match the structure of the obj parameter based on struct tags.
//
// Parameters:
//   - r: Pointer to the http.Request containing the JSON body.
//   - obj: Pointer to the object structure the JSON should map to.
//
// Returns:
//   - An error if the JSON is invalid or if it fails validation checks.
func (v *Validator) ValidateJson(r *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(obj)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return v.validate.Struct(obj)
}

// ValidateGET parses URL query parameters into a struct based on struct tags and performs validation.
// It supports conversion to string, int, and bool types based on the struct field types.
//
// Parameters:
//   - params: url.Values containing the URL query parameters.
//   - obj: Pointer to the object that should receive the values.
//
// Returns:
//   - An error if there is a mismatch in types or validation failures.
func (v *Validator) ValidateGET(params url.Values, obj interface{}) error {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Tag.Get("form")
		if fieldName == "" {
			fieldName = typ.Field(i).Name
		}

		paramValue := params.Get(fieldName)
		if paramValue == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(paramValue)
		case reflect.Int:
			if intValue, err := strconv.Atoi(paramValue); err == nil {
				field.SetInt(int64(intValue))
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(paramValue); err == nil {
				field.SetBool(boolValue)
			}
		}
	}

	return v.validate.Struct(obj)
}

// ValidatePOST parses form data from an HTTP POST request into a struct and performs validation.
// It supports similar type conversions and validations as ValidateGET.
//
// Parameters:
//   - r: Pointer to the http.Request from which form data is parsed.
//   - obj: Pointer to the object that should receive the form values.
//
// Returns:
//   - An error if the form data is invalid, cannot be parsed, or fails validation checks.
func (v *Validator) ValidatePOST(r *http.Request, obj interface{}) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid POST params: %w", err)
	}

	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Tag.Get("form")
		if fieldName == "" {
			fieldName = typ.Field(i).Name
		}

		paramValue := r.PostFormValue(fieldName)
		if paramValue == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(paramValue)
		case reflect.Int:
			if intValue, err := strconv.Atoi(paramValue); err == nil {
				field.SetInt(int64(intValue))
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(paramValue); err == nil {
				field.SetBool(boolValue)
			}
		}
	}

	return v.validate.Struct(obj)
}
