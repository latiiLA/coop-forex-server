package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TokenBlacklist struct {
	Token     string
	UserID    primitive.ObjectID
	IP        string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponseDTO struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenBlacklistRepository interface {
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time, userID primitive.ObjectID, ip string) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}
