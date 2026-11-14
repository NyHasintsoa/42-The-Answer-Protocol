package service

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func EnvChecker() error {
	_, err := ServerAddress()
	return err
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
