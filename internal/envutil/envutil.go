package envutil

import (
	"os"
	"strings"
)

func Getenv(key string) string {
	if filePath := os.Getenv(key + "_FILE"); filePath != "" {
		data, err := os.ReadFile(filePath)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return os.Getenv(key)
}
