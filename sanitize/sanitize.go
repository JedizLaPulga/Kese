package sanitize

import (
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// HTML escapes HTML special characters to prevent XSS attacks.
//
// Example:
//
//	safe := sanitize.HTML("<script>alert('xss')</script>")
//	// Returns: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
func HTML(input string) string {
	return html.EscapeString(input)
}

// SQL escapes single quotes to prevent SQL injection.
// Note: This is a basic escape. Prefer parameterized queries for real SQL safety.
//
// Example:
//
//	safe := sanitize.SQL("O'Reilly")
//	// Returns: "O''Reilly"
func SQL(input string) string {
	return strings.ReplaceAll(input, "'", "''")
}

// Path sanitizes file paths to prevent directory traversal attacks.
// Removes ".." and ensures the path stays within bounds.
//
// Example:
//
//	safe := sanitize.Path("../../etc/passwd")
//	// Returns: "etc/passwd"
func Path(input string) string {
	// Clean the path
	cleaned := filepath.Clean(input)

	// Remove leading slashes and parent directory references
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimPrefix(cleaned, "\\")

	// Remove any remaining ..
	for strings.Contains(cleaned, "..") {
		cleaned = strings.ReplaceAll(cleaned, "..", "")
	}

	return cleaned
}

// URL encodes a string for safe use in URLs.
//
// Example:
//
//	safe := sanitize.URL("hello world & foo=bar")
//	// Returns: "hello+world+%26+foo%3Dbar"
func URL(input string) string {
	return url.QueryEscape(input)
}

// AlphaNumeric removes all non-alphanumeric characters.
//
// Example:
//
//	safe := sanitize.AlphaNumeric("user-123!@#")
//	// Returns: "user123"
func AlphaNumeric(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(input, "")
}

// IsEmail validates if a string is a valid email format.
//
// Example:
//
//	valid := sanitize.IsEmail("user@example.com") // true
//	valid := sanitize.IsEmail("invalid-email")    // false
func IsEmail(email string) bool {
	// Simple email regex (not RFC 5322 compliant, but good enough)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsURL validates if a string is a valid URL format.
//
// Example:
//
//	valid := sanitize.IsURL("https://example.com") // true
//	valid := sanitize.IsURL("not-a-url")           // false
func IsURL(urlStr string) bool {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// StripTags removes all HTML/XML tags from a string.
//
// Example:
//
//	clean := sanitize.StripTags("<p>Hello <b>World</b></p>")
//	// Returns: "Hello World"
func StripTags(input string) string {
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	return tagRegex.ReplaceAllString(input, "")
}

// Truncate truncates a string to the specified length, adding "..." if truncated.
//
// Example:
//
//	short := sanitize.Truncate("This is a long string", 10)
//	// Returns: "This is a..."
func Truncate(input string, length int) string {
	if len(input) <= length {
		return input
	}
	return input[:length] + "..."
}
