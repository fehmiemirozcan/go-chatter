package command

import (
	"fmt"
	"strings"
)

func (h *Handler) CommandWhisper(outputChan chan string, args ...string) error {
	if len(args) < 2 {
		outputChan <- "Please provide a description that conforms to the specified format.[/whisper <user_name> : <message>]"
		return fmt.Errorf("invalid format")
	}

	input := strings.Join(args, " ")
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		outputChan <- "Please put a colon (:) between the username and the message."
		return fmt.Errorf("missing colon")
	}

	name := strings.TrimSpace(parts[0])
	message := strings.TrimSpace(parts[1])

	// Sadece ana bağlantıyı (ServerConn) kullanıyoruz
	if h.ServerConn == nil {
		outputChan <- "Error: You are not connected to any server to whisper!"
		return fmt.Errorf("no active connection")
	}

	msg := fmt.Sprintf("WHISPER|%s|%s\n", name, message)

	_, writeErr := h.ServerConn.Write([]byte(msg))
	if writeErr != nil {
		outputChan <- "Failed to send whisper."
		return fmt.Errorf("failed to send whisper")
	}

	whisperMsg := fmt.Sprintf("[whisper to %s]: %s", name, message)
	outputChan <- whisperMsg
	return nil
}
