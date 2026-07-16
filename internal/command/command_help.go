package command

import (
	"fmt"
	"strings"
)

func (h *Handler) CommandHelp(outputChan chan string, args ...string) error {

	var sb strings.Builder

	sb.WriteString("Commands you can use:\n")

	for _, cmd := range h.GetCommands() {
		msg := fmt.Sprintf("/%s - %s\n", cmd.name, cmd.description)
		sb.WriteString(msg)
	}

	outputChan <- sb.String()
	return nil
}
