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
			name:        "createroom",
			description: "<room_name> <room_password> ",
			Callback:    h.CommandCreateRoom,
		},
		"/join": {
			name:        "join",
			description: "<room_name> <room_password>",
			Callback:    h.CommandJoinRoom,
		},
		"/whisper": {
			name:        "whisper",
			description: "<user_name> : <message>",
			Callback:    h.CommandWhisper,
		},
		"/status": {
			name:        "status",
			description: "<online|away|dnd>",
			Callback:    h.CommandStatus,
		},
		"/quit": {
			name:        "quit",
			description: "Exiting the chat",
			Callback:    h.CommandExit,
		},
	}
}
