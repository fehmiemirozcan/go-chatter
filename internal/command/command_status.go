package command

import (
	"fmt"
	"strings"
)

func (h *Handler) CommandStatus(outputChan chan string, args ...string) error {
	if len(args) == 0 {
		outputChan <- "SYSTEM| Please specify a status. Usage: /status <online|away|dnd>"
		return nil
	}

	newStatus := strings.ToLower(args[0])

	if newStatus != "online" && newStatus != "away" && newStatus != "dnd" {
		outputChan <- "SYSTEM| Invalid status! Use: online, away, dnd"
		return nil
	}

	if h.ServerConn != nil {
		msg := fmt.Sprintf("STATUS|%s\n", newStatus)
		_, err := h.ServerConn.Write([]byte(msg))

		if err != nil {
			outputChan <- "SYSTEM| Error sending status to server."
		} else {
			outputChan <- fmt.Sprintf("SYSTEM| Status change request sent: %s", newStatus)
		}
	} else {
		outputChan <- "SYSTEM| Error: You are not connected to the server."
	}

	return nil
}
