package service

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// EnvChecker verifies that all required environment variables are present and
// non-empty.
//
// Returns:
//
//	An error naming any required variables that are missing or blank.
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

// ServerAddress reads SERVER_IP and SERVER_PORT from the environment and
// returns a host-port address after validating both values.
//
// Returns:
//
//	The joined host-port address, or an error if either environment value is invalid.
func ServerAddress() (string, error) {
	serverIP := strings.TrimSpace(os.Getenv("SERVER_IP"))
	serverPort := strings.TrimSpace(os.Getenv("SERVER_PORT"))

	port, err := strconv.Atoi(serverPort)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("SERVER_PORT must be an integer between 1 and 65535")
	}
	if net.ParseIP(serverIP) == nil {
		return "", fmt.Errorf("SERVER_IP must be a valid IP address")
	}

	return net.JoinHostPort(serverIP, serverPort), nil
}
