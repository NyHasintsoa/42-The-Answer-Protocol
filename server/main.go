package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	// Start listening on TCP port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Println("TCP Server started on port 8080. Waiting for connections...")

	for {
		// Accept incoming client connections
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each client concurrently in a new goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("New client connected: %s\n", clientAddr)

	// Send a welcome message to the client
	conn.Write([]byte("Connected to Go TCP Server! Type your message (type 'quit' to exit).\n"))

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Client %s disconnected or encountered error: %v\n", clientAddr, err)
			break
		}

		message = strings.TrimSpace(message)
		if message == "quit" {
			conn.Write([]byte("Goodbye!\n"))
			fmt.Printf("Client %s requested disconnection.\n", clientAddr)
			break
		}

		fmt.Printf("[%s] Received: %s\n", clientAddr, message)

		// Echo back the message with a timestamp
		response := fmt.Sprintf("Echo [%s]: %s\n", time.Now().Format("15:04:05"), message)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Printf("Error writing to client %s: %v\n", clientAddr, err)
			break
		}
	}
}
