package middleware

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
)

func JwtAuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := infrastructure.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			return
		}

		c.Set("userID", claims["userID"])
		c.Set("role", claims["role"])
		c.Set("branchID", claims["branchID"])
		c.Set("departmentID", claims["departmentID"])
		c.Set("permissions", claims["permissions"])
		c.Next()
	}
}

func AuthorizeRolesOrPermissions(allowedRoles []string, requiredPermission []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("inside auth role and permission")
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Role not found"})
			return
		}

		fmt.Print("required permissions", requiredPermission)
		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Invalid role format"})
			return
		}

		if slices.Contains(allowedRoles, role) {
			c.Next()
			return
		}

		perms, permsExist := c.Get("permissions")
		if !permsExist {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Permission not found"})
			return
		}

		permissions, ok := perms.([]interface{})
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Invalid permissions format"})
			return
		}

		userPerms := make([]string, 0, len(permissions))
		for _, p := range permissions {
			if str, ok := p.(string); ok {
				userPerms = append(userPerms, str)
			}
		}

		for _, required := range requiredPermission {
			if slices.Contains(userPerms, required) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access denied"})
	}
}
