package service

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ChatChannel int

const (
	ChannelGlobal ChatChannel = iota
	ChannelRoom
	ChannelGroup
)

func (c ChatChannel) String() string {
	switch c {
	case ChannelGlobal:
		return "Global"
	case ChannelRoom:
		return "Room"
	case ChannelGroup:
		return "Group"
	default:
		return "Chat"
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

func NewChatManager() *ChatManager {
	cm := &ChatManager{
		IsOpen:    false,
		ActiveTab: ChannelGlobal,
		Messages:  make(map[ChatChannel][]ChatMessage),
	}

	cm.Messages[ChannelGlobal] = []ChatMessage{
		{Sender: "Mike Busby", Text: "Yes I think so.", TimeStr: "3:21pm", Status: "Read 3:21pm", IsSelf: false, AvatarCol: rl.NewColor(100, 140, 200, 255)},
		{Sender: "You", Text: "What's your phone number? I have a question", TimeStr: "3:23pm", Status: "Read 3:23pm", IsSelf: true, AvatarCol: rl.NewColor(220, 150, 90, 255)},
		{Sender: "Mike Busby", Text: "It's 416-443-5299 and to answer your other question the item is in good condition.", TimeStr: "3:27pm", Status: "Read 3:27pm", IsSelf: false, AvatarCol: rl.NewColor(100, 140, 200, 255)},
		{Sender: "You", Text: "Awesome, I will call you in 20 mins when I am home", TimeStr: "3:48pm", Status: "Read 3:48pm", IsSelf: true, AvatarCol: rl.NewColor(220, 150, 90, 255)},
	}

	cm.Messages[ChannelRoom] = []ChatMessage{
		{Sender: "Party Leader", Text: "Welcome to the dungeon room!", TimeStr: "2:10pm", Status: "Read 2:10pm", IsSelf: false, AvatarCol: rl.NewColor(120, 180, 100, 255)},
	}

	cm.Messages[ChannelGroup] = []ChatMessage{
		{Sender: "Guild Bot", Text: "Guild raid scheduled for 8:00 PM.", TimeStr: "1:00pm", Status: "Read 1:00pm", IsSelf: false, AvatarCol: rl.NewColor(180, 100, 200, 255)},
	}

	return cm
}

func (cm *ChatManager) Toggle() {
	cm.IsOpen = !cm.IsOpen
}

func (cm *ChatManager) SendMessage(text string) {
	if text == "" {
		return
	}

	now := time.Now().Format("3:04pm")
	msg := ChatMessage{
		Sender:    "You",
		Text:      text,
		TimeStr:   now,
		Status:    fmt.Sprintf("Sent %s", now),
		IsSelf:    true,
		AvatarCol: rl.NewColor(220, 150, 90, 255),
	}

	cm.Messages[cm.ActiveTab] = append(cm.Messages[cm.ActiveTab], msg)
}

func (cm *ChatManager) GetCurrentMessages() []ChatMessage {
	return cm.Messages[cm.ActiveTab]
}