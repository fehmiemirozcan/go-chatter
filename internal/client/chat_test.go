package client

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Racemir/go-chatter/internal/command"
)

func TestStartChat_HostMigratin(t *testing.T) {
	var (
		originalReconnectDelay     = ReconnectDelay
		originalDiscoverLeaderFunc = DiscoverLeaderFunc
		originalStartUDPFunc       = StartUDPFunc
		originalDialFunc           = DialFunc
		originalStartServerFunc    = StartServerFunc
	)

	ReconnectDelay = 1 * time.Millisecond

	defer func() {
		ReconnectDelay = originalReconnectDelay
		DiscoverLeaderFunc = originalDiscoverLeaderFunc
		StartUDPFunc = originalStartUDPFunc
		DialFunc = originalDialFunc
		StartServerFunc = originalStartServerFunc
	}()

	tests := []struct {
		testName     string
		expectedChan string
	}{
		{"Yeni Lider Bulundu", "[!] New leader found at 192.168.1.100. Connecting..."},
		{"Lider Benim", "[!] You are the leader! Starting server..."},
	}

	for _, test := range tests {
		inputChan := make(chan string, 1)
		outputChan := make(chan string, 10)
		deathSignal := make(chan struct{})

		clientConn, serverConn := net.Pipe()
		serverConn.Close()

		StartUDPFunc = func(port string) {}
		StartServerFunc = func() {}

		if test.testName == "Lider Benim" {
			DiscoverLeaderFunc = func(port string) (string, error) {
				return "", fmt.Errorf("No one was found.")
			}
		} else {
			DiscoverLeaderFunc = func(port string) (string, error) {
				return "192.168.1.100", nil
			}
		}

		mockClient, mockServer := net.Pipe()
		DialFunc = func(network string, address string) (net.Conn, error) { return mockClient, nil }

		go func() {
			buf := make([]byte, 1024)
			mockServer.Read(buf)
			mockServer.Write([]byte("Succesful"))

		}()

		go StartChat("TestUser", clientConn, inputChan, outputChan, deathSignal)

		<-outputChan // "Welcome to the secret chat"
		<-outputChan // Help mesajı
		<-outputChan // Leader connection lost

		take := <-outputChan
		if !strings.Contains(take, test.expectedChan) {
			t.Errorf("[%s] Expected to contain: %q , Received: %q", test.testName, test.expectedChan, take)
		}

		mockClient.Close()
		mockServer.Close()
	}
}
func TestStartChat_InputAndDeathSignal(t *testing.T) {

	originalReconnectDelay := ReconnectDelay
	ReconnectDelay = 1 * time.Millisecond
	defer func() {
		ReconnectDelay = originalReconnectDelay
	}()
	tests := []struct {
		testName     string
		rawInput     string // klavyeden gönderilecek msg
		triggerDeath bool
		expectedChan string
	}{
		{"Boş Veri", " ", false, ""},
		{"Yanlış Komut", "/wrong_commant", false, "Unknown command. Type /help"},
		{"Normal mesaj", "selam", false, "ME| selam"},
		{"Death Signal", "", true, "Connection completely lost"},
	}

	for _, test := range tests {
		inputChan := make(chan string, 1)
		outputChan := make(chan string, 10)
		deathSignal := make(chan struct{})
		clientConn, serverConn := net.Pipe()

		go func() {
			buf := make([]byte, 1024)
			for {
				_, readErr := serverConn.Read(buf)
				if readErr != nil {
					break
				}
			}
		}()

		go StartChat("TestUser", clientConn, inputChan, outputChan, deathSignal)

		<-outputChan // "Welcome to the secret chat"
		<-outputChan // Help mesajı

		if test.triggerDeath == true {
			deathSignal <- struct{}{}
		} else {
			inputChan <- test.rawInput
		}

		if test.testName == "Boş Veri" {
			continue
		}

		take := <-outputChan
		if !strings.Contains(take, test.expectedChan) && take != test.expectedChan {
			t.Errorf("[%s] Expected to contain: %q, Received: %q", test.testName, test.expectedChan, take)
		}

		clientConn.Close()
		serverConn.Close()
	}
}
func TestParseIncomingMessage(t *testing.T) {
	// msg boş
	// msg başında ROOMCHANGE|SUCCESS|
	// msg başında ROOMCHANGE|ERR|
	// msg başında ALLROOMS|
	// msg başında ROOMUSERS|
	// msg başında WHISPER|
	// parts 2 ye eşitse yani msg 2 parçaya ayrılmışsa
	// mesaj içeriğinde : yoksa sistem mesajı olacak

	tests := []struct {
		testName string
		sent     string
		expected string //kanaldan beklenen mesaj

	}{
		{"Boş Veri", "", ""},
		{"Oda Değişikliği", "ROOMCHANGE|SUCCESS|", "[SYSTEM] Successfully moved to room ''!\n"},
		{"Hatalı Oda Değişikliği", "ROOMCHANGE|ERR|", "[ERROR] Room action failed: \n"},
		{"Oda Listesi", "ALLROOMS|", "ROOMS||"},
		{"Kullanıcı Listesi", "ROOMUSERS|", "USERS|"},
		{"Whisper Mesaj", "WHISPER|emir|merahaba", "WHISPER|emir|merahaba"},
		{"Msg 2 Ve ':' var", "emir: merhaba", "MSG| emir: merhaba\n"},
		{"Msg İçinde ':' Yok", "emir merhaba", "emir merhaba"},
	}

	for _, test := range tests {
		chanParsa := make(chan string, 1)
		cmdHand := &command.Handler{UserName: "emir"}

		parseIncomingMessage(test.sent, cmdHand, chanParsa)
		if test.expected != "" {
			take := <-chanParsa
			if take != test.expected {
				t.Errorf("Network Error ->> Expected:%q , Received:%q", test.expected, take)
			}
		}
	}
}

func TestCleanInput(t *testing.T) {
	// mesaj gönderip deneriz bu kısmı
	tests := []struct {
		sent         string
		Expected     []string
		inputControl bool
	}{
		{" /join  denemeler   ", []string{"/join", "denemeler"}, true},
		{" /join DeNeMeLeR", []string{"/join", "denemeler"}, true},
		{"/join  DeNeMeLeR   ", []string{"/join", "denemeler"}, true},
		{"   /jOIN     DeNeMeLeR  ", []string{"/join", "denemeler"}, true},
		{" /JOIN  denemeler", []string{"/join", "denemeler"}, true},
		{" /join  DENEMELER", []string{"/join", "denemeler"}, true},
	}

	for _, test := range tests {

		cleanOutput := cleanInput(test.sent)
		if !slices.Equal(cleanOutput, test.Expected) {
			t.Errorf("Expected:%v , Received:%v", test.Expected, cleanOutput)
		}
	}
}
func TestHandleUserCommand(t *testing.T) {

	// liste 0 ise
	// commandName /quit ise bağlanbtı kapat
	// commanddictionary false ise kanala "Unknown command. Type /help"
	tests := []struct {
		testName      string
		sent          string
		expected      string
		cloeseNetwork bool
	}{
		{"Boş Veri", "", "", false},
		{"/quit Geldiyse", "/quit", "", true}, // dikkat ilk deatlock
		{"/join Geldiyse", "/join", "Please specify a room name. Example: /join <room_name> <password>", false},
		{"Yanlış Komut Geldiyse", "/komut", "Unknown command. Type /help", false},
	}
	for _, test := range tests {
		cahnUserCom := make(chan string, 2)
		clientConn, serverConn := net.Pipe()

		hand := &command.Handler{ServerConn: serverConn, UserName: "TestUser"}

		handleUserCommand(test.sent, hand, cahnUserCom)

		if test.cloeseNetwork == true {
			_, writeErr := clientConn.Write([]byte("hello"))
			if writeErr == nil {
				t.Error("The /quit command was executed, but the network connection was not closed!")
			}
		}

		if test.expected != "" {
			take := <-cahnUserCom
			if test.expected != take {
				t.Errorf("Network Error ->> Expected:%q , Received:%q", test.expected, take)
			}
		}
	}
}

func TestSendNormalMessage(t *testing.T) {
	// saniye sınırı
	// bağlantı kopması
	// write bir sorunu
	// başarılı senaryo
	tests := []struct {
		testName           string
		sent               string
		expectedNetworkMsg string
		closeConn          bool
		spam               bool
	}{
		{"Spam Kontrolü", "merhaba", "SYSTEM| You're sending messages too quickly!", false, true},
		{"Bağlantı Kontrolü", "merhaba", "Error: You are not connected to the server.", false, false},
		{"Ağ Yazma Hatası", "merhaba", "Message could not be sent !!", true, false},
		{"Mesaj Gönderildi", "merhaba", "ME| merhaba", false, false},
	}
	for _, test := range tests {
		chanNormalMsg := make(chan string, 1)

		lastMsgTime := time.Now().Add(-1 * time.Hour)
		if test.spam == true {
			lastMsgTime = time.Now()
		}

		clientConn, serverConn := net.Pipe()

		if test.closeConn == true {
			serverConn.Close()
			clientConn.Close()
		}

		if test.testName == "Bağlantı Kontrolü" {
			serverConn = nil
		}

		if test.testName == "Mesaj Gönderildi" {
			go func() {
				buf := make([]byte, 1024)
				clientConn.Read(buf)

			}()
		}

		hand := &command.Handler{ServerConn: serverConn}
		sendNormalMessage(test.sent, hand, chanNormalMsg, &lastMsgTime)

		take := <-chanNormalMsg
		if take != test.expectedNetworkMsg {
			t.Errorf("Expected:%q , Received:%q", test.expectedNetworkMsg, take)
		}
	}
}
