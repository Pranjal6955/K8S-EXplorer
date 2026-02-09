package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Permission represents an action that can be performed
type Permission string

const (
	// Read permissions
	PermissionReadTopology     Permission = "read:topology"
	PermissionReadNodes        Permission = "read:nodes"
	PermissionReadEdges        Permission = "read:edges"
	PermissionReadDependencies Permission = "read:dependencies"

	// Write permissions
	PermissionWriteNodes Permission = "write:nodes"
	PermissionWriteEdges Permission = "write:edges"
	PermissionSync       Permission = "sync:cluster"

	// Admin permissions
	PermissionDelete    Permission = "delete:resources"
	PermissionClearData Permission = "clear:data"
	PermissionManage    Permission = "manage:all"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleViewer: {
		PermissionReadTopology,
		PermissionReadNodes,
		PermissionReadEdges,
		PermissionReadDependencies,
	},
	RoleOperator: {
		PermissionReadTopology,
		PermissionReadNodes,
		PermissionReadEdges,
		PermissionReadDependencies,
		PermissionWriteNodes,
		PermissionWriteEdges,
		PermissionSync,
	},
	RoleAdmin: {
		PermissionReadTopology,
		PermissionReadNodes,
		PermissionReadEdges,
		PermissionReadDependencies,
		PermissionWriteNodes,
		PermissionWriteEdges,
		PermissionSync,
		PermissionDelete,
		PermissionClearData,
		PermissionManage,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission || p == PermissionManage {
			return true
		}
	}

	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func HasAnyPermission(role Role, permissions ...Permission) bool {
	for _, p := range permissions {
		if HasPermission(role, p) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all specified permissions
func HasAllPermissions(role Role, permissions ...Permission) bool {
	for _, p := range permissions {
		if !HasPermission(role, p) {
			return false
		}
	}
	return true
}

// RequireRole middleware ensures the user has one of the specified roles
func RequireRole(roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c.Request.Context())
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "Insufficient permissions",
		})
		c.Abort()
	}
}

// RequirePermission middleware ensures the user has the specified permission
func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c.Request.Context())
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		if !HasPermission(user.Role, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Permission denied: " + string(permission),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission middleware ensures the user has at least one of the specified permissions
func RequireAnyPermission(permissions ...Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c.Request.Context())
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		if !HasAnyPermission(user.Role, permissions...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckNamespaceAccess checks if the user has access to the specified namespace
func CheckNamespaceAccess(user *User, namespace string) bool {
	// Admin role has access to all namespaces
	if user.Role == RoleAdmin {
		return true
	}

	// Empty namespace list means access to all namespaces
	if len(user.Namespace) == 0 {
		return true
	}

	// Check if namespace is in the allowed list
	for _, ns := range user.Namespace {
		if ns == namespace || ns == "*" {
			return true
		}
	}

	return false
}

// RequireNamespaceAccess middleware ensures the user has access to the namespace in the request
func RequireNamespaceAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c.Request.Context())
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get namespace from query or path
		namespace := c.Query("namespace")
		if namespace == "" {
			namespace = c.Param("namespace")
		}

		// If no namespace specified, allow (will be filtered later)
		if namespace == "" {
			c.Next()
			return
		}

		if !CheckNamespaceAccess(user, namespace) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Access denied for namespace: " + namespace,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAllowedNamespaces returns the namespaces a user has access to
// Returns nil if user has access to all namespaces
func GetAllowedNamespaces(user *User) []string {
	if user == nil {
		return []string{}
	}

	if user.Role == RoleAdmin {
		return nil // All namespaces
	}

	if len(user.Namespace) == 0 {
		return nil // All namespaces
	}

	for _, ns := range user.Namespace {
		if ns == "*" {
			return nil // All namespaces
		}
	}

	return user.Namespace
}
