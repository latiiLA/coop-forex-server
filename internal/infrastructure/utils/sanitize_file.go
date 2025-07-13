package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var filenameRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func SanitizeFilename(name string) string {
	// keep only base name
	base := filepath.Base(name)

	// replace unwanted runes
	cleaned := filenameRegex.ReplaceAllString(base, "_")

	// collapse repeated underscores
	cleaned = strings.Trim(cleaned, "_")
	cleaned = regexp.MustCompile(`_+`).ReplaceAllString(cleaned, "_")

	// collapse repeated underscores
	cleaned = strings.Trim(cleaned, "-")
	cleaned = regexp.MustCompile(`-+`).ReplaceAllString(cleaned, "_")

	// prevent leading dots / dashes
	cleaned = strings.TrimLeft(cleaned, "._-")

	// enforce length limit (e.g. 240 bytes)
	for len(cleaned) > 240 {
		cleaned = cleaned[:len(cleaned)-1]
	}

	// guard against Windows reserved names
	reserved := map[string]struct{}{
		"con": {}, "prn": {}, "aux": {}, "nul": {},
		"com1": {}, "lpt1": {},
	}
	lower := strings.ToLower(strings.TrimSuffix(cleaned, filepath.Ext(cleaned)))
	if _, bad := reserved[lower]; bad {
		cleaned = "_" + cleaned
	}

	return cleaned
}
