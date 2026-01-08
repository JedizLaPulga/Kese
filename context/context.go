package context

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/JedizLaPulga/kese/sanitize"
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

	// written tracks whether the response has been written
	written bool

	// bodyBytes stores the buffered request body for multiple reads
	bodyBytes []byte

	// bodyRead tracks whether the body has been read and buffered
	bodyRead bool

	// values stores arbitrary key-value pairs for passing data between middleware and handlers
	values map[string]interface{}
}

// New creates a new Context instance.
func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:    r,
		Writer:     w,
		params:     make(map[string]string),
		statusCode: http.StatusOK,
		written:    false,
		bodyBytes:  nil,
		bodyRead:   false,
		values:     make(map[string]interface{}),
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
// Returns the default value only if the parameter is not present in the URL.
// An empty string value ("?q=") will return empty string, not the default.
func (c *Context) QueryDefault(key, defaultValue string) string {
	values := c.Request.URL.Query()
	if !values.Has(key) {
		return defaultValue
	}
	return values.Get(key)
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
// Limited to 10MB to prevent memory exhaustion attacks.
// The body is buffered on first read, so this method can be called multiple times.
func (c *Context) Body(v interface{}) error {
	// Read and buffer the body if not already done
	if !c.bodyRead {
		defer c.Request.Body.Close()
		// Limit to 10MB to prevent memory exhaustion
		limitedReader := io.LimitReader(c.Request.Body, 10<<20) // 10 MB
		data, err := io.ReadAll(limitedReader)
		if err != nil {
			return err
		}
		c.bodyBytes = data
		c.bodyRead = true
	}

	// Parse JSON from buffered bytes
	return json.Unmarshal(c.bodyBytes, v)
}

// BodyBytes reads the raw request body as bytes.
// Limited to 10MB to prevent memory exhaustion attacks.
// The body is buffered on first read, so this method can be called multiple times.
func (c *Context) BodyBytes() ([]byte, error) {
	// Read and buffer the body if not already done
	if !c.bodyRead {
		defer c.Request.Body.Close()
		// Limit to 10MB to prevent memory exhaustion
		limitedReader := io.LimitReader(c.Request.Body, 10<<20) // 10 MB
		data, err := io.ReadAll(limitedReader)
		if err != nil {
			return nil, err
		}
		c.bodyBytes = data
		c.bodyRead = true
	}

	return c.bodyBytes, nil
}

// JSON sends a JSON response with the specified status code.
// The data will be marshaled to JSON automatically.
func (c *Context) JSON(status int, data interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.written = true

	encoder := json.NewEncoder(c.Writer)
	return encoder.Encode(data)
}

// JSONPretty sends a pretty-printed JSON response.
// Useful for debugging or human-readable APIs.
func (c *Context) JSONPretty(status int, data interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.written = true

	encoder := json.NewEncoder(c.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// String sends a plain text response.
func (c *Context) String(status int, text string) error {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.written = true

	_, err := c.Writer.Write([]byte(text))
	return err
}

// HTML sends an HTML response.
func (c *Context) HTML(status int, html string) error {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.written = true

	_, err := c.Writer.Write([]byte(html))
	return err
}

// Bytes sends a raw byte response with the specified content type.
func (c *Context) Bytes(status int, contentType string, data []byte) error {
	c.SetHeader("Content-Type", contentType)
	c.statusCode = status
	c.Writer.WriteHeader(c.statusCode)
	c.written = true

	_, err := c.Writer.Write(data)
	return err
}

// NoContent sends a 204 No Content response.
func (c *Context) NoContent() error {
	c.statusCode = http.StatusNoContent
	c.Writer.WriteHeader(http.StatusNoContent)
	c.written = true
	return nil
}

// Redirect sends a redirect response to the specified URL.
func (c *Context) Redirect(status int, url string) error {
	if status < 300 || status >= 400 {
		return fmt.Errorf("invalid redirect status code: %d (must be 3xx)", status)
	}
	http.Redirect(c.Writer, c.Request, url, status)
	c.written = true
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
	return c.written
}

// StatusCode returns the HTTP status code that was set.
func (c *Context) StatusCode() int {
	return c.statusCode
}

// SetWritten marks the response as written.
// This is used internally by static file serving and other methods that write directly.
func (c *Context) SetWritten() {
	c.written = true
}

// CSRFToken returns the CSRF token from context.
// Used in templates and handlers to access the current CSRF token.
func (c *Context) CSRFToken() string {
	if token := c.Get("csrf_token"); token != nil {
		return token.(string)
	}
	return ""
}

// Set stores a key-value pair in the context.
// This is useful for passing data between middleware and handlers.
// Example: c.Set("user", authenticatedUser)
func (c *Context) Set(key string, value interface{}) {
	c.values[key] = value
}

// Get retrieves a value from the context by key.
// Returns nil if the key doesn't exist.
// Example: user := c.Get("user")
func (c *Context) Get(key string) interface{} {
	return c.values[key]
}

// MustGet retrieves a value from the context and panics if it doesn't exist.
// Use this when you know the value must be present.
// Example: user := c.MustGet("user").(User)
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.values[key]; exists {
		return value
	}
	panic(fmt.Sprintf("key %q does not exist in context", key))
}

// Response helper methods for common HTTP status codes

// Success sends a 200 OK JSON response with the provided data.
func (c *Context) Success(data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

// Created sends a 201 Created JSON response with the provided resource.
func (c *Context) Created(resource interface{}) error {
	return c.JSON(http.StatusCreated, resource)
}

// BadRequest sends a 400 Bad Request JSON response with an error message.
func (c *Context) BadRequest(message string) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": message})
}

// Unauthorized sends a 401 Unauthorized JSON response.
func (c *Context) Unauthorized(message string) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": message})
}

// Forbidden sends a 403 Forbidden JSON response.
func (c *Context) Forbidden(message string) error {
	return c.JSON(http.StatusForbidden, map[string]string{"error": message})
}

// NotFoundError sends a 404 Not Found JSON response with an error message.
func (c *Context) NotFoundError(message string) error {
	return c.JSON(http.StatusNotFound, map[string]string{"error": message})
}

// InternalError sends a 500 Internal Server Error JSON response.
func (c *Context) InternalError(message string) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": message})
}

// Form data parsing methods

// FormValue returns the first value for the named form field from POST, PUT, or PATCH body.
// It calls ParseForm and ParseMultipartForm if necessary.
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// PostFormValue returns the first value for the named form field from POST, PUT, or PATCH body only.
// URL query parameters are ignored.
func (c *Context) PostFormValue(key string) string {
	return c.Request.PostFormValue(key)
}

// FormValues returns all values for the named form field.
func (c *Context) FormValues(key string) []string {
	c.Request.ParseForm()
	c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if values, ok := c.Request.Form[key]; ok {
		return values
	}
	return []string{}
}

// MultipartForm returns the multipart form data if the request is multipart/form-data.
// This is useful for accessing multiple form fields and files.
func (c *Context) MultipartForm() (*http.Request, error) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}
	return c.Request, nil
}

// File upload methods

// FormFile returns the first file for the provided form key.
// Returns the file, file header (with Filename, Size, etc.), and any error encountered.
func (c *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(key)
}

// SaveUploadedFile saves an uploaded file to the specified destination path.
// Example: c.SaveUploadedFile("avatar", "./uploads/avatar.png")
func (c *Context) SaveUploadedFile(formKey, dst string) error {
	file, _, err := c.Request.FormFile(formKey)
	if err != nil {
		return err
	}
	defer file.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	return err
}

// Sanitization helper methods

// SanitizeHTML escapes HTML to prevent XSS attacks.
func (c *Context) SanitizeHTML(input string) string {
	return sanitize.HTML(input)
}

// IsEmail validates if the input is a valid email.
func (c *Context) IsEmail(email string) bool {
	return sanitize.IsEmail(email)
}

// IsURL validates if the input is a valid URL.
func (c *Context) IsURL(urlStr string) bool {
	return sanitize.IsURL(urlStr)
}
