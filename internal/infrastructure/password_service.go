package infrastructure

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hash, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
	return err
}
