package infrastructure

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/latiiLA/coop-forex-server/configs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GenerateToken(userID primitive.ObjectID, role string, branchID primitive.ObjectID, departmentID primitive.ObjectID, permissions []string, ip string) (string, error) {
	claims := jwt.MapClaims{
		"userID":       userID.Hex(),
		"role":         role,
		"branchID":     branchID,
		"departmentID": departmentID,
		"permissions":  permissions,
		"ip":           ip,
		"exp":          time.Now().Add(time.Minute * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(configs.JwtSecret))
}

func ValidateToken(tokenString string, clientIP string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(configs.JwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("user ID missing in token")
	}

	// Check if user ID exists
	if _, exists := claims["userID"]; !exists {
		return nil, errors.New("user ID missing in token")
	}

	// --- IP validation ---
	tokenIP, ok := claims["ip"].(string)
	if !ok || tokenIP != clientIP {
		return nil, errors.New("IP does not match the token")
	}

	return claims, nil
}
