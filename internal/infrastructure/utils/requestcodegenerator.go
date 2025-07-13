package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateRequestCode() string {
	id := uuid.New().String()[:8]
	return fmt.Sprintf("REQ-%s", id)
}
