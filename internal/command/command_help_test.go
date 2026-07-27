package command

import (
	"strings"
	"testing"
)

func TestCommandHelp(t *testing.T) {
	commandList := []string{"/help", "/createroom", "/join", "/whisper", "/status", "/quit"}
	channel := make(chan string, 1)
	hand := &Handler{}

	cmdHelpErr := hand.CommandHelp(channel)
	if cmdHelpErr != nil {
		t.Fatalf("CommandHelp returned an unexpected error. %s", cmdHelpErr)
	}

	take := <-channel
	for _, cmd := range commandList {
		if !strings.Contains(take, cmd) {
			t.Errorf("Expected:%s data error", cmd)
		}
	}
}
