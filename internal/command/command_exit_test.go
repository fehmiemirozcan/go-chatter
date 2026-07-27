package command

import "testing"

func TestCommandExit(t *testing.T) {
	channel := make(chan string, 2)
	hand := &Handler{}

	quitErr := hand.CommandExit(channel)
	if quitErr != nil {
		t.Fatalf("CommandQuit returned an unexpected error. %s", quitErr)
	}

	text1 := "Closing the Chatter... Goodbye!"
	text2 := "SYS_EXIT_NOW"

	chanText := <-channel
	if text1 != chanText {
		t.Errorf("Expected:%q,Received:%q", text1, chanText)
	}

	chanText0 := <-channel
	if text2 != chanText0 {
		t.Errorf("Expected:%q,Received:%q", text2, chanText0)
	}
}
