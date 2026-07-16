package command

import (
	"net"
)

type Handler struct {
	UserName      string
	CurrentRoom   string // Kullanıcının o an aktif olarak mesaj gönderdiği oda.
	ServerConn    net.Conn
	RoomPasswords map[string]string // Odaların şifrelerini tutucam.
}

// ClientCommand struct
type ClientCommand struct {
	name        string
	description string
	Callback    func(outputChan chan string, args ...string) error
}

// getCommand fonskiyonu
func (h *Handler) GetCommands() map[string]ClientCommand {
	return map[string]ClientCommand{
		"/help": {
			name:        "help",
			description: "Displays a help message",
			Callback:    h.CommandHelp,
		},
		"/createroom": {
			name:        "createroom <room_name> <room_password> ",
			description: " Create a room",
			Callback:    h.CommandCreateRoom,
		},
		"/join": {
			name:        "join <room_name> <room_password>",
			description: "Join the room",
			Callback:    h.CommandJoinRoom,
		},
		"/whisper": {
			name:        "whisper <user_name> : <message>",
			description: "slient messge",
			Callback:    h.CommandWhisper,
		},
		"/status": {
			name:        "status <online|away|dnd>",
			description: "Change your presence status",
			Callback:    h.CommandStatus,
		},
		"/quit": {
			name:        "quit",
			description: "Exiting the chat",
			Callback:    h.CommandQuit,
		},
	}
}
