package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserID(c *gin.Context) (primitive.ObjectID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, errors.New("user ID not found in context")
	}

	userIDStr, ok := val.(string)
	if !ok {
		return primitive.NilObjectID, errors.New("user ID is not a valid string")
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid ObjectID format")
	}

	return userID, nil
}

func GetDepartmentID(c *gin.Context) (primitive.ObjectID, error) {
	val, exists := c.Get("departmentID")
	if !exists {
		return primitive.NilObjectID, errors.New("department ID not found in context")
	}

	departmentIDStr, ok := val.(string)
	if !ok {
		return primitive.NilObjectID, errors.New("branch ID is not a valid string")
	}

	departmentID, err := primitive.ObjectIDFromHex(departmentIDStr)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid ObjectID format")
	}

	return departmentID, nil
}

func GetBranchID(c *gin.Context) (primitive.ObjectID, error) {
	val, exists := c.Get("branchID")
	if !exists {
		return primitive.NilObjectID, errors.New("branch ID not found in context")
	}

	branchIDStr, ok := val.(string)
	if !ok {
		return primitive.NilObjectID, errors.New("branch ID is not a valid string")
	}

	branchID, err := primitive.ObjectIDFromHex(branchIDStr)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid ObjectID format")
	}

	return branchID, nil
}

func GetClaimIpAddress(c *gin.Context) (any, error) {
	val, exists := c.Get("ip")
	if !exists {
		return "", errors.New("ip address not found in context")
	}

	return val, nil
}

func GetIPAddress(c *gin.Context) (string, error) {
	// Use Gin ClientIP so X-Forwarded-For is only trusted for configured proxies.
	ip := c.ClientIP()
	if ip == "" {
		return "", errors.New("IP not found")
	}
	return ip, nil
}

func GetPermissions(c *gin.Context) []string {
	perms, exists := c.Get("permissions")
	if !exists {
		return nil
	}

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

func GetAccessToken(c *gin.Context) string {
	val, exists := c.Get("accessToken")
	if !exists {
		return ""
	}
	token, ok := val.(string)
	if !ok {
		return ""
	}
	return token
}
