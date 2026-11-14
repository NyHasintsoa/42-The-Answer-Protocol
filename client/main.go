// TAP terminal client for The Answer Protocol (RFC 42TAP).
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"tap-client/src/component"
	"tap-client/src/util"

	tea "charm.land/bubbletea/v2"
)

func main() {
	note := util.LoadEnv()
	ip := util.EnvOr("SERVER_IP", "127.0.0.1")
	port := util.EnvOr("SERVER_PORT", "8080")
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		fmt.Fprintf(os.Stderr, "Invalid SERVER_PORT %q (expected a number between 1 and 65535)\n", port)
		os.Exit(2)
	}

	addr := net.JoinHostPort(ip, port)
	program := tea.NewProgram(component.NewModel(addr, note+" - server "+addr))
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TAP CLI client: %v\n", err)
		os.Exit(1)
	}
}
