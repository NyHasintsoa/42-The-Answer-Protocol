package util

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnv reads ./.env without overriding variables already set in the environment.
func LoadEnv() string {
	switch err := godotenv.Load(); {
	case err == nil:
		return "Loaded .env"
	case errors.Is(err, os.ErrNotExist):
		return "No .env file found, using environment/defaults"
	default:
		return fmt.Sprintf("Could not read .env (%v), using environment/defaults", err)
	}
}

func EnvOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
