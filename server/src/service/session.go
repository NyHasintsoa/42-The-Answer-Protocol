package service

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"tap-server/src/model"
)

const maxMessageBytes = 1024

type CommandDispatcher interface {
	Dispatch(ClientSender, string) (string, bool)
	Disconnect(ClientSender)
	RecordSessionEvent(level, event, actor, remote, payload string)
	RecordCommand(actor, remote, command, parameters string)
}

type ClientSession struct {
	conn          net.Conn
	writeMu       sync.Mutex
	eventMu       sync.Mutex
	dispatching   bool
	pendingEvents []string
	player        *model.Player
	router        CommandDispatcher
}

func NewClientSession(conn net.Conn, router CommandDispatcher) *ClientSession {
	return &ClientSession{
		conn:   conn,
		player: &model.Player{},
		router: router,
	}
}

func (cs *ClientSession) SendLine(line string) {
	cs.eventMu.Lock()
	if cs.dispatching && strings.HasPrefix(line, "EVT ") {
		cs.pendingEvents = append(cs.pendingEvents, line)
		cs.eventMu.Unlock()
		return
	}
	cs.eventMu.Unlock()
	cs.writeLine(line)
}

func (cs *ClientSession) writeLine(line string) {
	cs.writeMu.Lock()
	defer cs.writeMu.Unlock()

	if err := cs.conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		cs.logWriteError(err, line)
		_ = cs.conn.Close()
		return
	}
	message := []byte(line + "\n")
	for len(message) > 0 {
		written, err := cs.conn.Write(message)
		if err != nil {
			cs.logWriteError(err, line)
			_ = cs.conn.Close()
			return
		}
		if written == 0 {
			err := errors.New("connection wrote zero bytes")
			cs.logWriteError(err, line)
			_ = cs.conn.Close()
			return
		}
		message = message[written:]
	}
}

func (cs *ClientSession) beginDispatch() {
	cs.eventMu.Lock()
	cs.dispatching = true
	cs.eventMu.Unlock()
}

func (cs *ClientSession) finishDispatch() {
	cs.eventMu.Lock()
	cs.dispatching = false
	for _, event := range cs.pendingEvents {
		cs.writeLine(event)
	}
	cs.pendingEvents = nil
	cs.eventMu.Unlock()
}

func (cs *ClientSession) GetPlayer() *model.Player {
	return cs.player
}

func (cs *ClientSession) Serve() {
	remote := cs.conn.RemoteAddr().String()
	cs.router.RecordSessionEvent("INFO", "client_connected", "", remote, "TCP connection established")
	defer func() {
		cs.router.Disconnect(cs)
		cs.router.RecordSessionEvent("INFO", "client_disconnected", cs.player.Username, remote, "TCP connection closed")
		_ = cs.conn.Close()
	}()

	cs.SendLine("OK hello proto=1")
	reader := bufio.NewReader(cs.conn)
	var windowStart time.Time
	commandCount := 0
	for {
		line, oversized, err := readProtocolLine(reader)
		if oversized {
			cs.router.RecordSessionEvent("WARN", "oversized_message", cs.player.Username, remote, fmt.Sprintf("limit_bytes=%d", maxMessageBytes))
			cs.router.RecordSessionEvent("WARN", "server_response", cs.player.Username, remote, `response="ERR 413 MESSAGE_TOO_LONG"`)
			cs.SendLine("ERR 413 MESSAGE_TOO_LONG")
		}
		if err != nil {
			return
		}
		if oversized {
			continue
		}
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		if !utf8.ValidString(line) {
			cs.router.RecordSessionEvent("WARN", "server_response", cs.player.Username, remote, `response="ERR 400 INVALID_ENCODING"`)
			cs.SendLine("ERR 400 INVALID_ENCODING")
			cs.router.RecordSessionEvent("WARN", "invalid_encoding", cs.player.Username, remote, "command was not valid UTF-8")
			continue
		}
		now := time.Now()
		if now.Sub(windowStart) >= time.Second {
			windowStart = now
			commandCount = 0
		}
		commandCount++
		if commandCount == 21 {
			cs.router.RecordSessionEvent("WARN", "command_flood", cs.player.Username, remote, "more than 20 commands in one second")
		}
		command, parameters, _ := strings.Cut(line, " ")
		actor := cs.player.Username
		if strings.EqualFold(command, "CONNECT") {
			actor = strings.TrimSpace(parameters)
		}
		cs.router.RecordCommand(actor, remote, command, parameters)
		cs.beginDispatch()
		response, shouldClose := cs.router.Dispatch(cs, line)
		cs.router.RecordSessionEvent(responseLevel(response), "server_response", cs.player.Username, remote, fmt.Sprintf("command=%q response=%q", line, response))
		cs.writeLine(response)
		cs.finishDispatch()
		if shouldClose {
			return
		}
	}
}

func readProtocolLine(reader *bufio.Reader) (string, bool, error) {
	var line []byte
	oversized := false
	for {
		fragment, err := reader.ReadSlice('\n')
		if !oversized {
			if len(line)+len(fragment) > maxMessageBytes {
				oversized = true
				line = nil
			} else {
				line = append(line, fragment...)
			}
		}
		if err == nil {
			return string(line), oversized, nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if errors.Is(err, io.EOF) {
			return "", oversized, io.EOF
		}
		return "", oversized, err
	}
}

func responseLevel(response string) string {
	if strings.HasPrefix(response, "ERR ") {
		return "WARN"
	}
	return "INFO"
}

func (cs *ClientSession) logWriteError(err error, line string) {
	if cs.router != nil {
		cs.router.RecordSessionEvent("ERROR", "send_failed", cs.player.Username, cs.conn.RemoteAddr().String(), fmt.Sprintf("response=%q error=%q", line, err.Error()))
	}
}
