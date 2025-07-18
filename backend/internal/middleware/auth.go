package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ringecosystem/degov-apps/internal/config"
)

// AuthClaims represents the JWT claims structure
type AuthClaims struct {
	Address string `json:"address"`
	jwt.RegisteredClaims
}

// ContextKey is used for context values to avoid conflicts
type ContextKey string

const (
	// UserClaimsKey is the context key for user claims
	UserClaimsKey ContextKey = "user_claims"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtSecret []byte
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware() *AuthMiddleware {
	secretKey := config.GetStringRequired("JWT_SECRET")
	return &AuthMiddleware{
		jwtSecret: []byte(secretKey),
	}
}

// HTTPMiddleware wraps HTTP handlers with JWT authentication
func (m *AuthMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		claims, err := m.validateToken(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			// You might want to return an error here depending on your requirements
			next.ServeHTTP(w, r)
			return
		}

		// Add user claims to context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// validateToken parses and validates a JWT token
func (m *AuthMiddleware) validateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	return claims, nil
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*AuthClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*AuthClaims)
	return claims, ok
}

// RequireAuth ensures user is authenticated, returns error if not
func RequireAuth(ctx context.Context) (*AuthClaims, error) {
	claims, ok := GetUserFromContext(ctx)
	if !ok || claims == nil {
		return nil, fmt.Errorf("authentication required")
	}
	return claims, nil
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(ctx context.Context) bool {
	_, ok := GetUserFromContext(ctx)
	return ok
}
