package utils_test

import (
	"testing"

	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
)

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		"normal-file.txt":          "normal_file.txt",
		"file with spaces.pdf":     "file_with_spaces.pdf",
		"file@#$%name!.zip":        "file_name_.zip",
		"../etc/passwd":            "passwd",
		"hello/../world":           "world",
		"#00":                      "00",
		"💥🔥🧨":                      "",          // all symbols replaced → cleaned = ""
		"CON.txt":                  "_CON.txt",  // reserved name
		"com1.exe":                 "_com1.exe", // reserved name
		"---hidden.file---":        "hidden.file",
		"some______file---name":    "some_file_name",
		"file___with---dashes.txt": "file_with_dashes.txt",
	}

	for input, expected := range cases {
		t.Run(input, func(t *testing.T) {
			actual := utils.SanitizeFilename(input)
			if actual != expected {
				t.Errorf("SanitizeFilename(%q) = %q; expected %q", input, actual, expected)
			}
		})
	}
}
