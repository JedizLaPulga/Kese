package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token has expired")
)

// Claims represents JWT claims (payload)
type Claims map[string]interface{}

// GenerateToken creates a new JWT token with the given claims.
//
// claims: Custom data to store in the token
// secret: Secret key for signing the token
// ttl: Time to live (e.g., 24*time.Hour)
//
// Example:
//
//	token, err := auth.GenerateToken(map[string]interface{}{
//	    "userID": "123",
//	    "email": "user@example.com",
//	}, "my-secret-key", 24*time.Hour)
func GenerateToken(claims Claims, secret string, ttl time.Duration) (string, error) {
	// Add standard claims
	now := time.Now()
	claims["iat"] = now.Unix()          // issued at
	claims["exp"] = now.Add(ttl).Unix() // expiration

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := createSignature(message, secret)

	// Combine parts
	token := message + "." + signature

	return token, nil
}

// ValidateToken validates a JWT token and returns its claims.
//
// Example:
//
//	claims, err := auth.ValidateToken(token, "my-secret-key")
//	if err != nil {
//	    // Invalid or expired token
//	}
//	userID := claims["userID"].(string)
func ValidateToken(token, secret string) (Claims, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	message := headerEncoded + "." + claimsEncoded
	expectedSignature := createSignature(message, secret)

	if signatureEncoded != expectedSignature {
		return nil, ErrInvalidToken
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrTokenExpired
		}
	}

	return claims, nil
}

// createSignature creates HMAC-SHA256 signature
func createSignature(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// RefreshToken creates a new token with the same claims but extended expiration.
// The original token must still be valid (not expired) to be refreshed.
// This prevents indefinite token refresh after expiration.
//
// Example:
//
//	newToken, err := auth.RefreshToken(oldToken, secret, 24*time.Hour)
func RefreshToken(token, secret string, ttl time.Duration) (string, error) {
	// Validate existing token - must not be expired
	claims, err := ValidateToken(token, secret)
	if err != nil {
		return "", err
	}

	// Remove old timestamps
	delete(claims, "iat")
	delete(claims, "exp")

	// Generate new token
	return GenerateToken(claims, secret, ttl)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header.
// Supports both "Bearer <token>" and just "<token>" formats.
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1], nil
	}

	if len(parts) == 1 {
		return parts[0], nil
	}

	return "", fmt.Errorf("invalid authorization header format")
}
