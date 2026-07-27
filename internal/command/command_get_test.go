package command

import (
	"testing"
)

func TestCommand(t *testing.T) {

	hand := &Handler{}

	commands := hand.GetCommands()
	// map'in içi dolu mu ?
	if commands == nil {
		t.Fatal("GetCommands returned an empty result.")
	}

	if len(commands) == 0 {
		t.Fatal("Getcommands is completely empty.")
	}

	commandList := []string{"/help", "/createroom", "/join", "/whisper", "/status", "/quit"}
	for _, cmd := range commandList {
		_, ok := commands[cmd]
		if !ok {
			t.Errorf("Expected command not found.")
		}
	}

	if len(commands) != len(commandList) {
		t.Errorf("Expected:%v, Received:%v", len(commands), len(commandList))
	}
}
