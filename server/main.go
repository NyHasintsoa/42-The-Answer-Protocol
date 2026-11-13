package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"tap-server/src/action"
	"tap-server/src/service"
	"tap-server/src/util"

	"github.com/joho/godotenv"
)

const (
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

// printCliError writes a categorized error to standard error using the provided color.
// category labels the error; details is the value to print; color is the terminal escape code.
func printCliError(category string, details any, color string) {
	fmt.Fprintf(os.Stderr, "%s[%s]%s %v\n", color, category, ColorReset, details)
}

const maxConcurrentClients = 256

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	listenAddress, err := service.ServerAddress()
	if err != nil {
		return err
	}
	store, err := util.NewJSONStore("data")
	if err != nil {
		return fmt.Errorf("initialize interaction store: %w", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Error("failed to close interaction store", "error", err)
		}
	}()

	worldSvc, err := service.NewWorldService("map.json")
	if err != nil {
		return fmt.Errorf("load map.json: %w", err)
	}

	broadcastSvc := service.NewBroadcastService()
	router := action.NewActionRouter(worldSvc, broadcastSvc, store)

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("bind TCP listener on %s: %w", listenAddress, err)
	}
	defer listener.Close()

	router.RecordSessionEvent("INFO", "server_started", "", "", "address="+listenAddress)

	slots := make(chan struct{}, maxConcurrentClients)
	var attemptsMu sync.Mutex
	recentAttempts := make(map[string][]time.Time)
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			logger.Error("failed to accept client connection", "error", err)
			continue
		}
		remote := conn.RemoteAddr().String()
		host, _, splitErr := net.SplitHostPort(remote)
		if splitErr != nil {
			host = remote
		}
		now := time.Now()
		attemptsMu.Lock()
		for address, recent := range recentAttempts {
			if len(recent) == 0 || now.Sub(recent[len(recent)-1]) > time.Minute {
				delete(recentAttempts, address)
			}
		}
		attempts := recentAttempts[host]
		firstRecent := 0
		for firstRecent < len(attempts) && now.Sub(attempts[firstRecent]) > time.Minute {
			firstRecent++
		}
		attempts = append(append([]time.Time(nil), attempts[firstRecent:]...), now)
		recentAttempts[host] = attempts
		rapid := len(attempts) > 10
		attemptsMu.Unlock()
		if rapid {
			router.RecordSessionEvent("WARN", "rapid_connections", "", remote, "more than 10 connection attempts in one minute")
		}
		select {
		case slots <- struct{}{}:
		default:
			router.RecordSessionEvent("WARN", "connection_limit", "", remote, fmt.Sprintf("maximum=%d", maxConcurrentClients))
			_ = conn.Close()
			continue
		}
		session := service.NewClientSession(conn, router)
		go func() {
			defer func() { <-slots }()
			session.Serve()
		}()
	}
}

// main starts the server and exits with an error status if startup fails.
func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		printCliError("Configuration Error", fmt.Errorf("load .env: %w", err), ColorRed)
		os.Exit(1)
	}
	if err := service.EnvChecker(); err != nil {
		printCliError("Configuration Error", err, ColorRed)
		os.Exit(1)
	}
	if err := run(); err != nil {
		printCliError("Unexpected Fatal Exception", err, ColorRed)
		os.Exit(1)
	}
}
