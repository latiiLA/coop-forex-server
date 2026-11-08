package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
)

// accessTokenBlacklist is set once at router bootstrap so logout can revoke access tokens
// without changing every JwtAuthMiddleware call site.
var accessTokenBlacklist model.TokenBlacklistRepository

func SetAccessTokenBlacklist(repo model.TokenBlacklistRepository) {
	accessTokenBlacklist = repo
}

func JwtAuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Status{Message: "Unauthorized or token expired", Error: "Unauthorized"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		clientIP, err := utils.GetIPAddress(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Status{Message: "Unauthorized or token expired", Error: "unauthorized"})
			return
		}

		claims, err := infrastructure.ValidateToken(token, clientIP)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Status{Message: "Unauthorized or token expired", Error: "Invalid token"})
			return
		}

		if accessTokenBlacklist != nil {
			blacklisted, blErr := accessTokenBlacklist.IsBlacklisted(c.Request.Context(), token)
			if blErr != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Status{Message: "Unauthorized or token expired", Error: "token check failed"})
				return
			}
			if blacklisted {
				c.AbortWithStatusJSON(http.StatusUnauthorized, response.Status{Message: "Unauthorized or token expired", Error: "token revoked"})
				return
			}
		}

		c.Set("userID", claims["userID"])
		c.Set("role", claims["role"])
		c.Set("branchID", claims["branchID"])
		c.Set("departmentID", claims["departmentID"])
		c.Set("ip", claims["ip"])
		c.Set("permissions", claims["permissions"])
		c.Set("accessToken", token)
		c.Next()
	}
}

func AuthorizeRolesOrPermissions(allowedRoles []string, requiredPermission []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logEntry := utils.GetLogger(c)
		logEntry.Info("inside auth role and permission")
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Status{Message: "Unauthorized or token expired", Error: "Role not found"})
			return
		}

		// security measures by adding ip to the token
		clientIP, err := utils.GetIPAddress(c)
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("invalid ip address")
			c.AbortWithStatusJSON(http.StatusBadRequest, response.Status{Message: "Unauthorized or token expired", Error: "cannot determine client IP"})
			return
		}

		tokenClaimIP, err := utils.GetClaimIpAddress(c)
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("invalid token ip address", tokenClaimIP)
			c.AbortWithStatusJSON(http.StatusBadRequest, response.Status{Message: "Unauthorized or token expired", Error: "cannot determine token IP"})
			return
		}

		if tokenClaimIP != clientIP {
			logEntry.Warn("trying from different ip")
			c.AbortWithStatusJSON(http.StatusBadRequest, response.Status{Message: "Unauthorized or token expired", Error: "your network has been changed. please relogin"})
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Status{Message: "Unauthorized or token expired", Error: "Invalid role format"})
			return
		}

		if slices.Contains(allowedRoles, role) {
			c.Next()
			return
		}

		perms, permsExist := c.Get("permissions")
		if !permsExist {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Status{Message: "Invalid permission", Error: "Permission not found"})
			return
		}

		userPerms := normalizePermissionClaim(perms)
		if userPerms == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Status{Message: "Invalid permission", Error: "Invalid permissions format"})
			return
		}

		for _, required := range requiredPermission {
			if slices.Contains(userPerms, required) {
				c.Next()
				return
			}
		}

		logEntry.Warn("message: Access Denied")
		c.AbortWithStatusJSON(http.StatusForbidden, response.Status{Message: "Access denied, invalid token", Error: "Access denied"})
	}
}

// normalizePermissionClaim accepts JWT permission claims as []string or []interface{}.
func normalizePermissionClaim(perms any) []string {
	switch v := perms.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, p := range v {
			if str, ok := p.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}
