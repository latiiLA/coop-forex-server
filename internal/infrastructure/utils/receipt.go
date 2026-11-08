package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

func PersonName(u *model.User) string {
	if u == nil || u.Profile == nil {
		return "N/A"
	}
	parts := []string{u.Profile.FirstName, u.Profile.MiddleName, u.Profile.LastName}
	out := make([]string, 0, 3)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return "N/A"
	}
	return strings.Join(out, " ")
}

func CurrencyLabel(c *model.Currency) string {
	if c == nil {
		return "N/A"
	}
	if c.ShortCode != "" {
		return fmt.Sprintf("%s (%s)", c.Name, c.ShortCode)
	}
	return c.Name
}

// GenerateReceiptNumber returns a unique receipt id, e.g. FH-RCP-20260916-A1B2C3.
func GenerateReceiptNumber() (string, error) {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("FH-RCP-%s-%s", time.Now().Format("20060102"), strings.ToUpper(hex.EncodeToString(buf))), nil
}
