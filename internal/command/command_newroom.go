package command

import (
	"fmt"
)

func (h *Handler) CommandCreateRoom(outputChan chan string, args ...string) error {
	return h.sendRoomAction(outputChan, "createroom", args...)
}

func (h *Handler) CommandJoinRoom(outputChan chan string, args ...string) error {
	return h.sendRoomAction(outputChan, "join", args...)
}

// Odaya katılmak ve oda kurmak için mevcut bağlantı üzerinden sinyal gönderir
func (h *Handler) sendRoomAction(outputChan chan string, action string, args ...string) error {
	if len(args) < 1 {
		outputChan <- fmt.Sprintf("Please specify a room name. Example: /%s <room_name> <password>", action)
		return fmt.Errorf("missing room name")
	}

	if len(args) > 2 {
		outputChan <- fmt.Sprintf("Error: Too many arguments. Example: /%s <room_name> [password]", action)
		return fmt.Errorf("too many arguments")
	}

	roomName := args[0]
	password := ""
	if len(args) > 1 {
		password = args[1]
	}

	// Mevcut ServerConn üzerinden komut yolluyoruz
	msg := fmt.Sprintf("CMD|%s|%s|%s\n", action, roomName, password)

	_, err := h.ServerConn.Write([]byte(msg))
	if err != nil {
		outputChan <- "Error: Could not send command to server."
		return err
	}

	// Şifreyi şimdiden kaydedelim ki mesajları çözebilelim
	h.RoomPasswords[roomName] = password

	return nil
}
