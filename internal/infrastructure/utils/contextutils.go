package utils

import (
	"errors"

	"strings"

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
	xff := c.Request.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0]) // first IP is the real client
		if ip != "" {
			return ip, nil
		}
	}

	// fallback to X-Real-IP
	ip := c.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip, nil
	}

	// fallback to remote address
	ip = c.ClientIP()
	if ip != "" {
		return ip, nil
	}

	return "", errors.New("IP not found")
}
