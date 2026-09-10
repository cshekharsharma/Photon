package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a new chiRouter with a new router instance.
func newTestRouter() *chiRouter {
	return &chiRouter{Router: chi.NewRouter()}
}

func TestRouterHTTPMethods(t *testing.T) {
	r := newTestRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Register route for each HTTP method
	methods := []struct {
		name     string
		function func(string, http.HandlerFunc)
	}{
		{"GET", r.Get},
		{"POST", r.Post},
		{"PUT", r.Put},
		{"DELETE", r.Delete},
		{"PATCH", r.Patch},
		{"HEAD", r.Head},
		{"OPTIONS", r.Options},
	}

	for _, method := range methods {
		method.function("/test", handler)

		req := httptest.NewRequest(method.name, "/test", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Handler should manage HTTP %s method", method.name)
	}
}

func TestMountAndHandle(t *testing.T) {
	r := newTestRouter()
	subRouter := chi.NewRouter()

	subRouter.Get("/sub", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Mount("/mount", subRouter)

	req := httptest.NewRequest("GET", "/mount/sub", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Mount should correctly handle subroutes")
}

func TestNotFoundAndMethodNotAllowedHandlers(t *testing.T) {
	r := newTestRouter()

	// Test Not Found Handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	// Test Method Not Allowed Handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Not Found
	reqNotFound := httptest.NewRequest("GET", "/nonexistent", nil)
	recNotFound := httptest.NewRecorder()
	r.ServeHTTP(recNotFound, reqNotFound)
	assert.Equal(t, http.StatusNotFound, recNotFound.Code, "NotFound handler should respond correctly")

	// Method Not Allowed
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	reqNotAllowed := httptest.NewRequest("POST", "/test", nil)
	recNotAllowed := httptest.NewRecorder()
	r.ServeHTTP(recNotAllowed, reqNotAllowed)
	assert.Equal(t, http.StatusMethodNotAllowed, recNotAllowed.Code, "MethodNotAllowed handler should respond correctly")
}

func TestUseMiddleware(t *testing.T) {
	r := newTestRouter()
	middlewareCalled := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	r.Use(middleware)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.True(t, middlewareCalled, "Middleware should be called")
	assert.Equal(t, http.StatusOK, rec.Code, "Middleware should allow request to proceed")
}

func TestRouterMethodAndMethodFunc(t *testing.T) {
	r := newTestRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test Method
	r.Method("GET", "/get", handler)
	req := httptest.NewRequest("GET", "/get", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "GET method should respond with OK")

	// Test MethodFunc
	r.MethodFunc("POST", "/post", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	reqPost := httptest.NewRequest("POST", "/post", nil)
	recPost := httptest.NewRecorder()
	r.ServeHTTP(recPost, reqPost)
	assert.Equal(t, http.StatusOK, recPost.Code, "POST method should respond with OK")
}

func TestRouterHandleAndHandleFunc(t *testing.T) {
	r := newTestRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test Handle
	r.Handle("/handle", handler)
	reqHandle := httptest.NewRequest("GET", "/handle", nil)
	recHandle := httptest.NewRecorder()
	r.ServeHTTP(recHandle, reqHandle)
	assert.Equal(t, http.StatusOK, recHandle.Code, "Handle should respond to GET with OK")

	// Test HandleFunc
	r.HandleFunc("/handlefunc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	reqFunc := httptest.NewRequest("POST", "/handlefunc", nil)
	recFunc := httptest.NewRecorder()
	r.ServeHTTP(recFunc, reqFunc)
	assert.Equal(t, http.StatusOK, recFunc.Code, "HandleFunc should respond to POST with OK")
}
