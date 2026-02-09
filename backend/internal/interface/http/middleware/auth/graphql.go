package auth

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

// GraphQLAuthDirective implements the @auth directive for GraphQL
type GraphQLAuthDirective struct {
	jwtService *JWTService
}

// NewGraphQLAuthDirective creates a new auth directive handler
func NewGraphQLAuthDirective(jwtService *JWTService) *GraphQLAuthDirective {
	return &GraphQLAuthDirective{jwtService: jwtService}
}

// RequireAuth is a directive that requires authentication
func (d *GraphQLAuthDirective) RequireAuth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil, &AuthError{Message: "Authentication required"}
	}
	return next(ctx)
}

// RequireRole is a directive that requires a specific role
func (d *GraphQLAuthDirective) RequireRole(ctx context.Context, obj interface{}, next graphql.Resolver, role Role) (interface{}, error) {
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil, &AuthError{Message: "Authentication required"}
	}

	if user.Role != role && user.Role != RoleAdmin {
		return nil, &ForbiddenError{Message: "Insufficient permissions: requires role " + string(role)}
	}

	return next(ctx)
}

// RequirePermission is a directive that requires a specific permission
func (d *GraphQLAuthDirective) RequirePermission(ctx context.Context, obj interface{}, next graphql.Resolver, permission Permission) (interface{}, error) {
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil, &AuthError{Message: "Authentication required"}
	}

	if !HasPermission(user.Role, permission) {
		return nil, &ForbiddenError{Message: "Permission denied: " + string(permission)}
	}

	return next(ctx)
}

// RequireNamespace is a directive that checks namespace access
func (d *GraphQLAuthDirective) RequireNamespace(ctx context.Context, obj interface{}, next graphql.Resolver, namespace string) (interface{}, error) {
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil, &AuthError{Message: "Authentication required"}
	}

	if !CheckNamespaceAccess(user, namespace) {
		return nil, &ForbiddenError{Message: "Access denied for namespace: " + namespace}
	}

	return next(ctx)
}

// AuthError represents an authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// ForbiddenError represents an authorization error
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// GraphQLAuth provides authentication helpers for GraphQL resolvers
type GraphQLAuth struct{}

// NewGraphQLAuth creates a new GraphQL auth helper
func NewGraphQLAuth() *GraphQLAuth {
	return &GraphQLAuth{}
}

// CheckAuth checks if the user is authenticated
func (a *GraphQLAuth) CheckAuth(ctx context.Context) (*User, error) {
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil, &AuthError{Message: "Authentication required"}
	}
	return user, nil
}

// CheckRole checks if the user has the required role
func (a *GraphQLAuth) CheckRole(ctx context.Context, requiredRoles ...Role) (*User, error) {
	user, err := a.CheckAuth(ctx)
	if err != nil {
		return nil, err
	}

	for _, role := range requiredRoles {
		if user.Role == role {
			return user, nil
		}
	}

	// Admin always has access
	if user.Role == RoleAdmin {
		return user, nil
	}

	return nil, &ForbiddenError{Message: "Insufficient permissions"}
}

// CheckPermission checks if the user has the required permission
func (a *GraphQLAuth) CheckPermission(ctx context.Context, permission Permission) (*User, error) {
	user, err := a.CheckAuth(ctx)
	if err != nil {
		return nil, err
	}

	if !HasPermission(user.Role, permission) {
		return nil, &ForbiddenError{Message: "Permission denied: " + string(permission)}
	}

	return user, nil
}

// CheckNamespace checks if the user has access to the namespace
func (a *GraphQLAuth) CheckNamespace(ctx context.Context, namespace string) (*User, error) {
	user, err := a.CheckAuth(ctx)
	if err != nil {
		return nil, err
	}

	if !CheckNamespaceAccess(user, namespace) {
		return nil, &ForbiddenError{Message: "Access denied for namespace: " + namespace}
	}

	return user, nil
}

// FilterNamespaces filters a list of namespaces based on user access
func (a *GraphQLAuth) FilterNamespaces(ctx context.Context, namespaces []string) []string {
	user := GetUserFromContext(ctx)
	if user == nil {
		return []string{}
	}

	allowed := GetAllowedNamespaces(user)
	if allowed == nil {
		return namespaces // All namespaces allowed
	}

	// Create a set of allowed namespaces
	allowedSet := make(map[string]bool)
	for _, ns := range allowed {
		allowedSet[ns] = true
	}

	// Filter namespaces
	result := make([]string, 0)
	for _, ns := range namespaces {
		if allowedSet[ns] {
			result = append(result, ns)
		}
	}

	return result
}
