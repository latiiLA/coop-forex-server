package middleware

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
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
		clientIP, err := utils.GetIPAddress(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := infrastructure.ValidateToken(token, clientIP)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			return
		}

		c.Set("userID", claims["userID"])
		c.Set("role", claims["role"])
		c.Set("branchID", claims["branchID"])
		c.Set("departmentID", claims["departmentID"])
		c.Set("ip", claims["ip"])
		c.Set("permissions", claims["permissions"])
		c.Next()
	}
}

func AuthorizeRolesOrPermissions(allowedRoles []string, requiredPermission []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logEntry := utils.GetLogger(c)
		logEntry.Info("inside auth role and permission")
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Role not found"})
			return
		}

		// security measures by adding ip to the token
		clientIP, err := utils.GetIPAddress(c)
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("invalid ip address")
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "cannot determine client IP"})
			return
		}

		tokenClaimIP, err := utils.GetClaimIpAddress(c)
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("invalid token ip address", tokenClaimIP)
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "cannot determine token IP"})
			return
		}

		if tokenClaimIP != clientIP {
			logEntry.Warn("trying from different ip")
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "your network has been changed. please relogin"})
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

		logEntry.Warn("message: Access Denied")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access denied"})
	}
}
