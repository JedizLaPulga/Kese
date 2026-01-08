package kese

import (
	"fmt"
	"net/http"

	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/health"
	"github.com/JedizLaPulga/kese/logger"
	"github.com/JedizLaPulga/kese/router"
)

// HandlerFunc defines the function signature for route handlers.
// It receives a Context and returns an error for centralized error handling.
type HandlerFunc func(*context.Context) error

// App is the main application instance that holds the router and configuration.
// It provides a high-level API for defining routes and middleware.
type App struct {
	router         *router.Router
	middleware     []MiddlewareFunc
	errorHandler   ErrorHandler
	healthCheck    *health.HealthChecker
	Logger         *logger.Logger
	templateEngine *TemplateEngine
}

// MiddlewareFunc defines the function signature for middleware.
// Middleware can modify the context or terminate the request chain.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// New creates a new Kese application instance.
// This is the starting point for building your web application.
func New() *App {
	return &App{
		router:       router.New(),
		middleware:   make([]MiddlewareFunc, 0),
		errorHandler: DefaultErrorHandler,
		healthCheck:  health.New(),
		Logger:       logger.New(),
	}
}

// Use adds middleware to the application.
// Middleware is executed in the order it is registered.
func (a *App) Use(middleware ...MiddlewareFunc) {
	a.middleware = append(a.middleware, middleware...)
}

// SetErrorHandler sets a custom error handler for the application.
// The error handler receives errors from route handlers and returns appropriate responses.
func (a *App) SetErrorHandler(handler ErrorHandler) {
	a.errorHandler = handler
}

// SetTemplateEngine sets the template engine for rendering HTML templates.
// After calling this, use app.RenderTemplate() in handlers to render templates.
//
// Example:
//
//	engine := kese.NewTemplateEngine("./templates")
//	engine.LoadTemplates("*.html")
//	app.SetTemplateEngine(engine)
func (a *App) SetTemplateEngine(engine *TemplateEngine) {
	a.templateEngine = engine
}

// RenderTemplate renders an HTML template using the app's template engine.
// The template engine must be set via SetTemplateEngine first.
//
// Example:
//
//	func handler(c *context.Context) error {
//	    data := map[string]interface{}{"Title": "Home"}
//	    return app.RenderTemplate(c, 200, "home.html", data)
//	}
func (a *App) RenderTemplate(c *context.Context, status int, name string, data interface{}) error {
	if a.templateEngine == nil {
		return c.InternalError("Template engine not set. Call SetTemplateEngine first.")
	}
	return a.templateEngine.Render(c, status, name, data)
}

func (a *App) AddHealthCheck(name string, check health.CheckFunc) {
	a.healthCheck.AddCheck(name, check)
}

// HealthHandler returns the health check HTTP handler.
func (a *App) HealthHandler() HandlerFunc {
	return func(c *context.Context) error {
		a.healthCheck.ServeHTTP(c.Writer, c.Request)
		c.SetWritten()
		return nil
	}
}

// GET registers a route that responds to GET requests.
func (a *App) GET(path string, handler HandlerFunc) {
	a.addRoute(http.MethodGet, path, handler)
}

// POST registers a route that responds to POST requests.
func (a *App) POST(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPost, path, handler)
}

// PUT registers a route that responds to PUT requests.
func (a *App) PUT(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPut, path, handler)
}

// DELETE registers a route that responds to DELETE requests.
func (a *App) DELETE(path string, handler HandlerFunc) {
	a.addRoute(http.MethodDelete, path, handler)
}

// PATCH registers a route that responds to PATCH requests.
func (a *App) PATCH(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPatch, path, handler)
}

// OPTIONS registers a route that responds to OPTIONS requests.
func (a *App) OPTIONS(path string, handler HandlerFunc) {
	a.addRoute(http.MethodOptions, path, handler)
}

// HEAD registers a route that responds to HEAD requests.
func (a *App) HEAD(path string, handler HandlerFunc) {
	a.addRoute(http.MethodHead, path, handler)
}

// addRoute is the internal method for registering routes with the router.
func (a *App) addRoute(method, path string, handler HandlerFunc) {
	// Wrap the handler with all registered middleware
	wrappedHandler := a.wrapMiddleware(handler)
	a.router.Add(method, path, wrappedHandler)
}

// wrapMiddleware wraps a handler with all registered middleware.
// Middleware is applied in reverse order so that the first registered
// middleware is the outermost layer.
func (a *App) wrapMiddleware(handler HandlerFunc) HandlerFunc {
	// Apply middleware in reverse order
	for i := len(a.middleware) - 1; i >= 0; i-- {
		handler = a.middleware[i](handler)
	}
	return handler
}

// RouterGroup represents a group of routes with a common prefix and middleware.
type RouterGroup struct {
	app        *App
	prefix     string
	middleware []MiddlewareFunc
}

// Group creates a new router group with the given prefix and optional middleware.
// Example: api := app.Group("/api/v1", authMiddleware())
func (a *App) Group(prefix string, middleware ...MiddlewareFunc) *RouterGroup {
	return &RouterGroup{
		app:        a,
		prefix:     prefix,
		middleware: middleware,
	}
}

// GET registers a GET route within the group.
func (rg *RouterGroup) GET(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodGet, path, handler)
}

// POST registers a POST route within the group.
func (rg *RouterGroup) POST(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodPost, path, handler)
}

// PUT registers a PUT route within the group.
func (rg *RouterGroup) PUT(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodPut, path, handler)
}

// DELETE registers a DELETE route within the group.
func (rg *RouterGroup) DELETE(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodDelete, path, handler)
}

// PATCH registers a PATCH route within the group.
func (rg *RouterGroup) PATCH(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodPatch, path, handler)
}

// OPTIONS registers an OPTIONS route within the group.
func (rg *RouterGroup) OPTIONS(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodOptions, path, handler)
}

// HEAD registers a HEAD route within the group.
func (rg *RouterGroup) HEAD(path string, handler HandlerFunc) {
	rg.addRoute(http.MethodHead, path, handler)
}

// addRoute adds a route to the app with the group's prefix and middleware.
func (rg *RouterGroup) addRoute(method, path string, handler HandlerFunc) {
	// Apply group's middleware to the handler
	for i := len(rg.middleware) - 1; i >= 0; i-- {
		handler = rg.middleware[i](handler)
	}

	// Add the route to the main app with the prefixed path
	fullPath := rg.prefix + path
	rg.app.addRoute(method, fullPath, handler)
}

// ServeHTTP implements http.Handler interface.
// This allows the App to be used directly with http.Server.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a new context for this request
	ctx := context.New(w, r)

	// Find the matching route
	handlerInterface, params := a.router.Match(r.Method, r.URL.Path)
	if handlerInterface == nil {
		// No route matched - return 404
		ctx.String(http.StatusNotFound, "404 Not Found")
		return
	}

	// Type assert the handler from interface{} to HandlerFunc
	handler, ok := handlerInterface.(HandlerFunc)
	if !ok {
		// This should never happen if we're using the framework correctly
		ctx.String(http.StatusInternalServerError, "Internal Error: invalid handler type")
		return
	}

	// Set route parameters in context
	ctx.SetParams(params)

	// Execute the handler
	if err := handler(ctx); err != nil {
		// Handle errors returned by handlers using the custom error handler
		// Only write error response if no response has been written yet
		if !ctx.IsWritten() {
			statusCode, response := a.errorHandler(err)
			ctx.JSON(statusCode, response)
		}
		// If response was already written, we can't send error info to client
		// but we could log it here if needed
	}

}

// Run starts the HTTP server on the specified address.
// address should be in the format ":8080" or "localhost:8080"
func (a *App) Run(address string) error {
	fmt.Printf("🚀 Kese server starting on %s\n", address)
	return http.ListenAndServe(address, a)
}

// RunTLS starts the HTTPS server on the specified address with TLS config.
func (a *App) RunTLS(address, certFile, keyFile string) error {
	fmt.Printf("🔒 Kese server starting on %s (TLS)\n", address)
	return http.ListenAndServeTLS(address, certFile, keyFile, a)
}
