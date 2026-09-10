package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ChiRouter is a wrapper for the Chi Mux router, implementing the http.Handler interface.
type chiRouter struct {
	Router *chi.Mux
}

// ServeHTTP implements the http.Handler interface, allowing ChiRouter to serve HTTP requests.
// It dispatches incoming HTTP requests to the appropriate route registered in the router.
func (r *chiRouter) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	r.Router.ServeHTTP(w, request)
}

// Mount attaches a sub-router under the specified pattern.
// It associates the provided http.Handler with the given pattern, making it available under that path.
func (r *chiRouter) Mount(pattern string, handler http.Handler) {
	r.Router.Mount(pattern, handler)
}

// Handle registers a new route with the given pattern and handler.
// It associates the provided http.Handler with the specified URL pattern, using the default HTTP methods.
func (r *chiRouter) Handle(pattern string, handler http.Handler) {
	r.Router.Handle(pattern, handler)
}

// HandleFunc registers a new route with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern, using the default HTTP methods.
func (r *chiRouter) HandleFunc(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Handle(pattern, handlerFn)
}

// Method registers a new route for a specific HTTP method, pattern, and handler.
// It associates the provided http.Handler with the given URL pattern and HTTP method.
func (r *chiRouter) Method(method string, pattern string, handler http.Handler) {
	r.Router.Method(method, pattern, handler)
}

// MethodFunc registers a new route for a specific HTTP method, pattern, and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern and HTTP method.
func (r *chiRouter) MethodFunc(method, pattern string, handlerFunc http.HandlerFunc) {
	r.Router.MethodFunc(method, pattern, handlerFunc)
}

// Head registers a new route for the HTTP HEAD method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the HEAD method.
func (r *chiRouter) Head(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Head(pattern, handlerFn)
}

// Options registers a new route for the HTTP OPTIONS method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the OPTIONS method.
func (r *chiRouter) Options(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Options(pattern, handlerFn)
}

// Get registers a new route for the HTTP GET method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the GET method.
func (r *chiRouter) Get(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Get(pattern, handlerFn)
}

// Post registers a new route for the HTTP POST method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the POST method.

func (r *chiRouter) Post(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Post(pattern, handlerFn)
}

// Put registers a new route for the HTTP PUT method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the PUT method.
func (r *chiRouter) Put(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Put(pattern, handlerFn)

}

// Delete registers a new route for the HTTP DELETE method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the DELETE method.
func (r *chiRouter) Delete(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Delete(pattern, handlerFn)

}

// Patch registers a new route for the HTTP PATCH method with the given pattern and http.HandlerFunc.
// It associates the provided http.HandlerFunc with the specified URL pattern for the PATCH method.
func (r *chiRouter) Patch(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Patch(pattern, handlerFn)
}

// NotFound sets the handler function for routes that are not found.
// It associates the provided http.HandlerFunc with routes that match no other registered routes.
func (r *chiRouter) NotFound(handlerFn http.HandlerFunc) {
	r.Router.NotFound(handlerFn)
}

// MethodNotAllowed sets the handler function for routes that have a method not allowed.
// It associates the provided http.HandlerFunc with routes that match, but the HTTP method is not allowed.
func (r *chiRouter) MethodNotAllowed(handlerFn http.HandlerFunc) {
	r.Router.MethodNotAllowed(handlerFn)
}

// Use applies one or more middleware functions to all routes within the router.
// Middleware functions are used for preprocessing requests or postprocessing responses.
func (r *chiRouter) Use(middlewares ...func(http.Handler) http.Handler) {
	r.Router.Use(middlewares...)
}
