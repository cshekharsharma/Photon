package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/stretchr/testify/assert"
)

func TestGetLocale(t *testing.T) {
	t.Run("GetLocaleFromContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), LocaleKey, "en-US")
		locale := GetLocale(ctx)
		assert.Equal(t, "en-US", locale)
	})

	t.Run("GetLocaleFromEmptyContext", func(t *testing.T) {
		ctx := context.Background()
		assert.Panics(t, func() {
			GetLocale(ctx)
		})
	})
}

func TestSetLocale(t *testing.T) {
	t.Run("SetLocaleInRequestContext", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req = SetLocale(req, "en-US")
		locale := req.Context().Value(LocaleKey).(string)
		assert.Equal(t, "en-US", locale)
	})
}

func TestRequestInit(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := r.Context().Value(LocaleKey).(string)
		if _, err := w.Write([]byte(locale)); err != nil {
			t.Fatalf("failed to write locale response: %v", err)
		}
	})

	middleware := RequestInit(nextHandler)

	t.Run("RequestWithLanguageHeader", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(rest.HeaderContentLanguage, "en-US")
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, "en-US", rr.Body.String())
	})

	t.Run("RequestWithoutLanguageHeader", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, "", rr.Body.String())
	})

	t.Run("RequestWithInvalidHeader", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(rest.HeaderContentLanguage, "invalid")
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, "", rr.Body.String())
	})
}

// RequestInit must leave r.Body readable by downstream handlers so webhook
// signature verification (which needs the raw payload) still works. Prior to
// the body-buffering fix, r.ParseForm() would drain the body for form-encoded
// POSTs and the handler would receive an empty stream.
func TestRequestInit_BodyRemainsReadableAfterParseForm(t *testing.T) {
	const payload = "signature=abc&payment_status=COMPLETE&amount_gross=10.00"

	var downstreamBody string
	var downstreamForm string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		downstreamBody = string(body)
		downstreamForm = r.PostFormValue("signature")
	})

	req, _ := http.NewRequest("POST", "/", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	RequestInit(next).ServeHTTP(rr, req)

	assert.Equal(t, payload, downstreamBody, "raw body must survive ParseForm")
	assert.Equal(t, "abc", downstreamForm, "PostForm must be populated for handlers that rely on it")
}

// A GET request has no body — RequestInit must still populate the locale
// context and not panic on the nil-body branch.
func TestRequestInit_NilBodyGET(t *testing.T) {
	var got string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Context().Value(LocaleKey).(string)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Body = nil
	rr := httptest.NewRecorder()

	RequestInit(next).ServeHTTP(rr, req)

	assert.Equal(t, "", got)
}

type requestInitErrReadCloser struct{}

func (requestInitErrReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("parse failed")
}

func (requestInitErrReadCloser) Close() error {
	return nil
}

func TestParseForm_IgnoresParseError(t *testing.T) {
	req := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body: requestInitErrReadCloser{},
	}

	parseForm(req)
}
