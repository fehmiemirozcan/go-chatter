package command

import "time"

func (h *Handler) CommandQuit(outputChan chan string, args ...string) error {

	outputChan <- "Closing the Chatter... Goodbye!"

	time.Sleep(500 * time.Millisecond)

	outputChan <- "SYS_EXIT_NOW"

	return nil
}
