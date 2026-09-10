package rest

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define test structs
type TestJsonRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=130"`
}

func TestValidateJson(t *testing.T) {
	validate := New()
	validJson := `{"name":"John Doe","email":"john.doe@example.com","age":30}`
	invalidJson := `{"name":"","email":"not-an-email","ag`

	t.Run("valid JSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/", bytes.NewBufferString(validJson))
		require.NoError(t, err)
		req.Header.Set(HeaderContentType, ContentTypeJSON)

		var request TestJsonRequest
		err = validate.ValidateJson(req, &request)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/", bytes.NewBufferString(invalidJson))
		require.NoError(t, err)
		req.Header.Set(HeaderContentType, ContentTypeJSON)

		var request TestJsonRequest
		err = validate.ValidateJson(req, &request)
		assert.Error(t, err)
	})
}

func TestValidator_ValidateGET(t *testing.T) {
	type TestStruct struct {
		Name     string `form:"name" validate:"required"`
		Age      int    `form:"age"`
		Verified bool   `form:"verified"`
		Role     string `validate:"required"`
	}

	cases := []struct {
		name         string
		params       url.Values
		expectError  bool
		expectedData TestStruct
	}{
		{
			name: "Valid inputs",
			params: url.Values{
				"name":     {"Alice"},
				"age":      {"30"},
				"verified": {"true"},
				"Role":     {"admin"},
			},
			expectError: false,
			expectedData: TestStruct{
				Name:     "Alice",
				Age:      30,
				Verified: true,
				Role:     "admin",
			},
		},
		{
			name: "Missing optional fields",
			params: url.Values{
				"name": {"Bob"},
				"Role": {"user"},
			},
			expectError: false,
			expectedData: TestStruct{
				Name: "Bob",
				Role: "user",
			},
		},
		{
			name: "Invalid int and bool",
			params: url.Values{
				"name":     {"John"},
				"age":      {"invalid"},
				"verified": {"notabool"},
				"Role":     {"guest"},
			},
			expectError: false,
			expectedData: TestStruct{
				Name: "John",
				Role: "guest",
			},
		},
		{
			name:         "Missing required field",
			params:       url.Values{},
			expectError:  true,
			expectedData: TestStruct{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &TestStruct{}
			v := &Validator{validate: validator.New()}

			err := v.ValidateGET(tc.params, obj)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, *obj)
			}
		})
	}
}

func TestValidatePOST(t *testing.T) {
	type testForm struct {
		Name  string `form:"name" validate:"required"`
		Age   int    `form:"age"`
		Admin bool   `form:"admin"`
		Role  string `validate:"required"`
	}

	tests := []struct {
		name           string
		form           url.Values
		expectError    bool
		expectedName   string
		expectedAge    int
		expectedAdmin  bool
		prepareReq     func() *http.Request
		setupValidator func() *Validator
	}{
		{
			name: "valid form data",
			form: url.Values{
				"name":  {"Alice"},
				"age":   {"30"},
				"admin": {"true"},
				"Role":  {"admin"},
			},
			expectedName:  "Alice",
			expectedAge:   30,
			expectedAdmin: true,
			expectError:   false,
			prepareReq: func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader("name=Alice&age=30&admin=true&Role=admin"))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			setupValidator: func() *Validator {
				return &Validator{validate: validator.New()}
			},
		},
		{
			name: "missing required field",
			form: url.Values{
				"age":   {"25"},
				"admin": {"false"},
				"":      {"Bob"},
			},
			expectError: true,
			prepareReq: func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader("age=25&admin=false"))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			setupValidator: func() *Validator {
				return &Validator{validate: validator.New()}
			},
		},
		{
			name: "invalid integer value",
			form: url.Values{
				"name": {"Bob"},
				"age":  {"invalid"},
				"Role": {"user"},
			},
			expectedName: "Bob",
			expectedAge:  0, // should remain 0
			expectError:  false,
			prepareReq: func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader("name=Bob&age=invalid&Role=user"))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			setupValidator: func() *Validator {
				return &Validator{validate: validator.New()}
			},
		},
		{
			name:        "ParseForm returns error",
			expectError: true,
			prepareReq: func() *http.Request {
				r := &http.Request{
					Method: "POST",
					Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
					Body:   io.NopCloser(brokenReader{}), // simulate read error
				}
				return r
			},
			setupValidator: func() *Validator {
				return &Validator{validate: validator.New()}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &testForm{}
			req := tt.prepareReq()
			validator := tt.setupValidator()
			err := validator.ValidatePOST(req, form)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, form.Name)
				assert.Equal(t, tt.expectedAge, form.Age)
				assert.Equal(t, tt.expectedAdmin, form.Admin)
			}
		})
	}
}

type brokenReader struct{}

func (brokenReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

func (brokenReader) Close() error {
	return nil
}
