package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ROUTER_CHI is a constant that represents the Chi router.
const RouterCHI = "chi"

// Router is an interface that defines the routing mechanisms for HTTP requests.
// It embeds the http.Handler interface and provides methods for registering routes,
// handling HTTP methods, mounting sub-routers, and configuring middleware.
type Router interface {
	http.Handler

	// ServeHTTP dispatches the request to the handler registered with the router.
	// It implements the http.Handler interface.
	// Parameters:
	//  - w: http.ResponseWriter, used to write responses to the client.
	//  - request: *http.Request, represents the HTTP request received.
	ServeHTTP(w http.ResponseWriter, request *http.Request)

	// Mount attaches a http.Handler at the specified pattern in the Router.
	// It is a wrapper around the Mux' Mount method, providing a way to
	// easily integrate with the routing mechanism. This method is useful
	// for mounting sub-routers or adding prefixed routes to the current router.
	//
	// Usage:
	//  - pattern: a URL pattern to match. It can include patterns like '/articles/:id'
	//    where ':id' is a variable part of the URL.
	//  - handler: the http.Handler to be mounted at the specified pattern.
	//
	// This method allows for modular and clean routing setups, especially useful
	// in larger applications where routes can be grouped and managed separately.
	//
	// Example:
	//  // Assuming Router is already created and initialized
	//  r := router.NewRouter()
	//
	//  // Creating a sub-router
	//  subRouter := router.NewRouter()
	//  subRouter.Get("/users/{userID}", getUserHandler)
	//  subRouter.Post("/users", createUserHandler)
	//
	//  // Mounting the sub-router on the main router
	//  r.Mount("/api", subRouter)
	//
	// Note: It's important to ensure that the patterns in the sub-router don't
	// conflict with those in the main router, as the main router's patterns
	// take precedence.
	//
	// The library provides extensive features for routing, including middleware
	// support, URL parameters, and more. Using Mount in conjunction with these
	// features can help in building a robust and scalable web application.
	Mount(pattern string, handler http.Handler)

	// Handle registers a new route with a specific pattern and a handler.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handler: http.Handler, the handler to execute for the route.
	Handle(pattern string, handler http.Handler)

	// HandleFunc registers a route with a pattern and an http.HandlerFunc.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	HandleFunc(pattern string, handlerFn http.HandlerFunc)

	// Method registers a route for a specific HTTP method, pattern, and handler.
	// Parameters:
	//  - method: string, the HTTP method (GET, POST, etc.).
	//  - pattern: string, the URL pattern to match.
	//  - handler: http.Handler, the handler to execute for the route.
	Method(method, pattern string, handler http.Handler)

	// MethodFunc registers a route for a specific HTTP method, pattern, and http.HandlerFunc.
	// Parameters:
	//  - method: string, the HTTP method (GET, POST, etc.).
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	MethodFunc(method, pattern string, handlerFn http.HandlerFunc)

	// Head registers a new route for the HTTP HEAD method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Head(pattern string, handlerFn http.HandlerFunc)

	// Options registers a new route for the HTTP OPTIONS method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Options(pattern string, handlerFn http.HandlerFunc)

	// Get registers a new route for the HTTP GET method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Get(pattern string, handlerFn http.HandlerFunc)

	// Post registers a new route for the HTTP POST method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Post(pattern string, handlerFn http.HandlerFunc)

	// Put registers a new route for the HTTP PUT method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Put(pattern string, handlerFn http.HandlerFunc)

	// Delete registers a new route for the HTTP DELETE method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Delete(pattern string, handlerFn http.HandlerFunc)

	// Patch registers a new route for the HTTP PATCH method.
	// Parameters:
	//  - pattern: string, the URL pattern to match.
	//  - handlerFn: http.HandlerFunc, the function to execute for the route.
	Patch(pattern string, handlerFn http.HandlerFunc)

	// NotFound sets the handler function for routes that are not found.
	// This is used to customize the response for unmatched routes.
	// Parameters:
	//  - handlerFn: http.HandlerFunc, the function to execute when no route is matched.
	NotFound(handlerFn http.HandlerFunc)

	// MethodNotAllowed sets the handler function for routes that have a method not allowed.
	// This is used to customize the response when a route exists but does not support the
	// requested HTTP method.
	// Parameters:
	//  - handlerFn: http.HandlerFunc, the function to execute when the method is not allowed.
	MethodNotAllowed(handlerFn http.HandlerFunc)

	// Use applies one or more middleware functions to all routes within the router.
	// Middleware is used for processing requests before reaching the final handler or after
	// the response has been written.
	// Parameters:
	//  - middlewares: variadic parameter, accepting multiple middleware functions.
	// Each middleware is a function that takes and returns an http.Handler.
	Use(middlewares ...func(http.Handler) http.Handler)
}

// NewRouter creates a new Router based on the specified provider.
// It returns a router instance of the specified type, or nil if the provider is not supported.
// NewRouter creates a new Router based on the specified provider.
// It accepts a provider string that indicates the type of router to create.
// Supported providers include ROUTER_CHI for Chi router.
//
// If the provided provider string matches ROUTER_CHI, it returns a new ChiRouter instance as a Router.
// If the provider is not supported or empty, it returns nil.
//
// Parameters:
//   - provider (string): A string representing the router provider (e.g., ROUTER_CHI).
//
// Returns:
//   - Router: A Router instance corresponding to the specified provider, or nil if the provider is not supported.
func NewRouter(provider string) Router {
	if provider == RouterCHI {
		return NewChiRouter()
	}

	return nil
}

// NewChiRouter creates a new ChiRouter instance and returns it as a Router.
//
// Returns:
//   - Router: A Router instance representing a Chi router.
func NewChiRouter() Router {
	chi := chi.NewRouter()
	chiRouter := &chiRouter{
		Router: chi,
	}

	return chiRouter
}
