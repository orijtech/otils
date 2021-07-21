package otils

import (
	"os"
	"strings"
)

func EnvOrAlternates(envVar string, alternates ...string) string {
	if retr := strings.TrimSpace(os.Getenv(envVar)); retr != "" {
		return retr
	}
	for _, alt := range alternates {
		alt = strings.TrimSpace(alt)
		if alt != "" {
			return alt
		}
	}
	return ""
}
