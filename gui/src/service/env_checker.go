package service

import (
	"fmt"
	"os"
	"strings"
)

func EnvChecker() error {
	requiredVars := []string{"SERVER_IP", "SERVER_PORT"}
	var missing []string

	for _, v := range requiredVars {
		val, exists := os.LookupEnv(v)
		if !exists || strings.TrimSpace(val) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or empty required environment variables '%s'", strings.Join(missing, ", "))
	}
	return nil
}