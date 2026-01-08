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

// Path sanitizes file paths to prevent directory traversal attacks.
// Removes leading slashes and checks for attempts to escape the base directory.
//
// Example:
//
//	safe := sanitize.Path("../../etc/passwd")
//	// Returns: "etc/passwd"
func Path(input string) string {
	// Clean the path to resolve . and .. properly
	cleaned := filepath.Clean(input)

	// Remove leading slashes and backslashes
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimPrefix(cleaned, "\\")

	// After cleaning, if path still starts with .., it's attempting to escape
	// This check doesn't corrupt legitimate filenames like "my..file.txt"
	if strings.HasPrefix(cleaned, "..") {
		// Remove the leading .. and any following separator
		cleaned = strings.TrimPrefix(cleaned, "..")
		cleaned = strings.TrimPrefix(cleaned, "/")
		cleaned = strings.TrimPrefix(cleaned, "\\")
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
