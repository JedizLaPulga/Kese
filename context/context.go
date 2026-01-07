package context

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Context wraps http.Request and http.ResponseWriter to provide
// a convenient API for handling requests and sending responses.
// It provides high-level methods while still exposing the underlying
// Request and Writer for advanced use cases.
type Context struct {
	// Request is the underlying HTTP request.
	// Exposed for advanced use cases where developers need direct access.
	Request *http.Request

	// Writer is the underlying HTTP response writer.
	// Exposed for advanced use cases where developers need direct access.
	Writer http.ResponseWriter

	// params stores URL path parameters extracted by the router
	params map[string]string

	// statusCode tracks the HTTP status code that was set
	statusCode int

	// Written tracks whether the response has been written
	Written bool
}

// New creates a new Context instance.
func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:    r,
		Writer:     w,
		params:     make(map[string]string),
		statusCode: http.StatusOK,
		Written:    false,
	}
}

// SetParams sets the route parameters for this context.
// This is called by the router after matching a route.
func (c *Context) SetParams(params map[string]string) {
	c.params = params
}

// Param returns the value of a URL path parameter.
// For example, for the route "/users/:id", Param("id") returns the ID value.
func (c *Context) Param(key string) string {
	return c.params[key]
}

// Query returns the value of a URL query parameter.
// For example, for the URL "/search?q=golang", Query("q") returns "golang".
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault returns the value of a URL query parameter with a default fallback.
func (c *Context) QueryDefault(key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Header returns the value of a request header.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header.
// This must be called before writing the response body.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Status sets the HTTP status code for the response.
// The status code will be written when a response method is called.
func (c *Context) Status(code int) {
	c.statusCode = code
}

// Body parses the request body as JSON into the provided value.
// The value should be a pointer to a struct.
func (c *Context) Body(v interface{}) error {
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(v)
}

// BodyBytes reads the raw request body as bytes.
func (c *Context) BodyBytes() ([]byte, error) {
	defer c.Request.Body.Close()
	return io.ReadAll(c.Request.Body)
}

// JSON sends a JSON response with the specified status code.
// The data will be marshaled to JSON automatically.
func (c *Context) JSON(status int, data interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.Written = true

	encoder := json.NewEncoder(c.Writer)
	return encoder.Encode(data)
}

// JSONPretty sends a pretty-printed JSON response.
// Useful for debugging or human-readable APIs.
func (c *Context) JSONPretty(status int, data interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.Written = true

	encoder := json.NewEncoder(c.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// String sends a plain text response.
func (c *Context) String(status int, text string) error {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.Written = true

	_, err := c.Writer.Write([]byte(text))
	return err
}

// HTML sends an HTML response.
func (c *Context) HTML(status int, html string) error {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.Written = true

	_, err := c.Writer.Write([]byte(html))
	return err
}

// Bytes sends a raw byte response with the specified content type.
func (c *Context) Bytes(status int, contentType string, data []byte) error {
	c.SetHeader("Content-Type", contentType)
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.Written = true

	_, err := c.Writer.Write(data)
	return err
}

// NoContent sends a 204 No Content response.
func (c *Context) NoContent() error {
	c.Writer.WriteHeader(http.StatusNoContent)
	c.Written = true
	return nil
}

// Redirect sends a redirect response to the specified URL.
func (c *Context) Redirect(status int, url string) error {
	if status < 300 || status >= 400 {
		return fmt.Errorf("invalid redirect status code: %d (must be 3xx)", status)
	}
	http.Redirect(c.Writer, c.Request, url, status)
	c.Written = true
	return nil
}

// Cookie returns the named cookie from the request.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie adds a Set-Cookie header to the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// Method returns the HTTP method of the request.
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the path of the request URL.
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// IsWritten returns true if the response has been written.
func (c *Context) IsWritten() bool {
	return c.Written
}

// StatusCode returns the HTTP status code that was set.
func (c *Context) StatusCode() int {
	return c.statusCode
}
