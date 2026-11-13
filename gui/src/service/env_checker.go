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