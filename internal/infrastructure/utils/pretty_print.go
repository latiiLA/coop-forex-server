package utils

import (
	"encoding/json"
	"fmt"
)

func pretty(v any) {
	out, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(out))
}
