package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Role represents a user role
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// User represents an authenticated user
type User struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Role      Role     `json:"role"`
	Namespace []string `json:"namespaces,omitempty"` // Allowed namespaces (empty = all)
}

// Claims represents JWT claims
type Claims struct {
	UserID     string   `json:"user_id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Role       Role     `json:"role"`
	Namespaces []string `json:"namespaces,omitempty"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string
	Issuer        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:     "your-secret-key-change-in-production",
		Issuer:        "k8s-graph-explorer",
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
}

// JWTService handles JWT operations
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{config: config}
}

// GenerateToken generates a new JWT token for a user
func (s *JWTService) GenerateToken(user *User) (string, error) {
	claims := &Claims{
		UserID:     user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		Namespaces: user.Namespace,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateRefreshToken generates a refresh token
func (s *JWTService) GenerateRefreshToken(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractTokenFromHeader extracts the token from Authorization header
func ExtractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// Context keys
type contextKey string

const (
	userContextKey   contextKey = "user"
	claimsContextKey contextKey = "claims"
)

// GetUserFromContext retrieves the user from context
func GetUserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return nil
}

// GetClaimsFromContext retrieves the claims from context
func GetClaimsFromContext(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(claimsContextKey).(*Claims); ok {
		return claims
	}
	return nil
}

// SetUserInContext sets the user in context
func SetUserInContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// SetClaimsInContext sets the claims in context
func SetClaimsInContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// JWTMiddleware creates a Gin middleware for JWT authentication
func JWTMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ExtractTokenFromHeader(c.Request)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing authorization token",
			})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Create user from claims
		user := &User{
			ID:        claims.UserID,
			Username:  claims.Username,
			Email:     claims.Email,
			Role:      claims.Role,
			Namespace: claims.Namespaces,
		}

		// Set user and claims in context
		ctx := SetUserInContext(c.Request.Context(), user)
		ctx = SetClaimsInContext(ctx, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalJWTMiddleware creates a middleware that allows unauthenticated requests
func OptionalJWTMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ExtractTokenFromHeader(c.Request)
		if token == "" {
			c.Next()
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			// Invalid token but still allow the request
			c.Next()
			return
		}

		user := &User{
			ID:        claims.UserID,
			Username:  claims.Username,
			Email:     claims.Email,
			Role:      claims.Role,
			Namespace: claims.Namespaces,
		}

		ctx := SetUserInContext(c.Request.Context(), user)
		ctx = SetClaimsInContext(ctx, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
