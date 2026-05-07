package controllers

import (
	"net/http"

	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/middleware/auth"
	"github.com/gin-gonic/gin"
)

// AuthController handles authentication endpoints
type AuthController struct {
	jwtService *auth.JWTService
}

// NewAuthController creates a new AuthController
func NewAuthController(jwtService *auth.JWTService) *AuthController {
	return &AuthController{
		jwtService: jwtService,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresIn    int64     `json:"expiresIn"`
	User         *UserInfo `json:"user"`
}

// UserInfo represents user information in response
type UserInfo struct {
	ID         string   `json:"id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Role       string   `json:"role"`
	Namespaces []string `json:"namespaces,omitempty"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Login handles user login
// @Summary      User login
// @Description  Authenticates a user and returns JWT tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  LoginRequest  true  "Login credentials"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// TODO: Replace with actual user authentication
	// This is a demo implementation
	user := c.authenticateUser(req.Username, req.Password)
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid username or password",
		})
		return
	}

	// Generate tokens
	accessToken, err := c.jwtService.GenerateToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate access token",
		})
		return
	}

	refreshToken, err := c.jwtService.GenerateRefreshToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate refresh token",
		})
		return
	}

	ctx.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds
		User: &UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       string(user.Role),
			Namespaces: user.Namespace,
		},
	})
}

// RefreshToken refreshes an access token
// @Summary      Refresh token
// @Description  Refreshes an access token using a refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  RefreshRequest  true  "Refresh token"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate refresh token
	claims, err := c.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or expired refresh token",
		})
		return
	}

	// Create user from claims
	user := &auth.User{
		ID:        claims.UserID,
		Username:  claims.Username,
		Email:     claims.Email,
		Role:      claims.Role,
		Namespace: claims.Namespaces,
	}

	// Generate new access token
	accessToken, err := c.jwtService.GenerateToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate access token",
		})
		return
	}

	ctx.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Return same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    86400,
		User: &UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       string(user.Role),
			Namespaces: user.Namespace,
		},
	})
}

// GetCurrentUser returns the current authenticated user
// @Summary      Get current user
// @Description  Returns the currently authenticated user's information
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserInfo
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/me [get]
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	user := auth.GetUserFromContext(ctx.Request.Context())
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
		return
	}

	ctx.JSON(http.StatusOK, UserInfo{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Role:       string(user.Role),
		Namespaces: user.Namespace,
	})
}

// Logout handles user logout
// @Summary      User logout
// @Description  Logs out the current user (client should discard tokens)
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse
// @Router       /api/v1/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// JWT is stateless, so logout is handled client-side
	// In a production system, you might want to:
	// 1. Add the token to a blacklist
	// 2. Use short-lived tokens with refresh tokens
	// 3. Store token state in Redis

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
	})
}

// authenticateUser authenticates a user (demo implementation)
// Replace this with actual user authentication logic
func (c *AuthController) authenticateUser(username, password string) *auth.User {
	// Demo users for testing
	demoUsers := map[string]*auth.User{
		"admin": {
			ID:        "1",
			Username:  "admin",
			Email:     "admin@example.com",
			Role:      auth.RoleAdmin,
			Namespace: []string{}, // All namespaces
		},
		"operator": {
			ID:        "2",
			Username:  "operator",
			Email:     "operator@example.com",
			Role:      auth.RoleOperator,
			Namespace: []string{"default", "kube-system"},
		},
		"viewer": {
			ID:        "3",
			Username:  "viewer",
			Email:     "viewer@example.com",
			Role:      auth.RoleViewer,
			Namespace: []string{"default"},
		},
	}

	// Demo: accept any password for demo users
	if user, exists := demoUsers[username]; exists && password != "" {
		return user
	}

	return nil
}
