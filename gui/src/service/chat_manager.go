package service

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ChatChannel int

const (
	ChannelGlobal ChatChannel = iota
	ChannelRoom
	ChannelGroup
)

// String returns the display name of c, defaulting to "Global" for unknown channels.
func (c ChatChannel) String() string {
	switch c {
	case ChannelRoom:
		return "Room"
	case ChannelGroup:
		return "Group"
	default:
		return "Global"
	}
}

type ChatMessage struct {
	Sender    string
	Text      string
	TimeStr   string
	Status    string
	IsSelf    bool
	AvatarCol rl.Color
}

type ChatManager struct {
	IsOpen    bool
	ActiveTab ChatChannel
	Messages  map[ChatChannel][]ChatMessage
}

// NewChatManager creates an empty chat manager.
func NewChatManager() *ChatManager {
	cm := &ChatManager{
		IsOpen:    false,
		ActiveTab: ChannelGlobal,
		Messages:  make(map[ChatChannel][]ChatMessage),
	}
	return cm
}

// Toggle changes whether the chat panel is open.
func (cm *ChatManager) Toggle() {
	cm.IsOpen = !cm.IsOpen
}

// AddMessage appends a server-reported chat message to the selected channel.
func (cm *ChatManager) AddMessage(channel ChatChannel, sender, text string, isSelf bool) {
	avatar := rl.NewColor(100, 140, 200, 255)
	if isSelf {
		avatar = rl.NewColor(220, 150, 90, 255)
	}
	cm.Messages[channel] = append(cm.Messages[channel], ChatMessage{
		Sender: sender, Text: text, IsSelf: isSelf,
		TimeStr: time.Now().Format("3:04pm"), Status: "Received",
		AvatarCol: avatar,
	})
}

// GetCurrentMessages returns the messages in the active channel.
func (cm *ChatManager) GetCurrentMessages() []ChatMessage {
	return cm.Messages[cm.ActiveTab]
}
