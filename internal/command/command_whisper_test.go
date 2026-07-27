package command

import (
	"net"
	"testing"
)

func TestCommandWhisper(t *testing.T) {

	// gelen liste uzunluğu 2 olmalı
	// gelen listeyi ":" işaretle 2 parçaya bölmeli
	// mesaj doğruı olup bağlantıda sorun olması
	// gönderilen mesaj doğru mu ?
	// kanala giden mesaj doğru mu ?

	tests := []struct {
		testName           string
		sent               []string
		expected           string // kanaldan beklediğim msg
		expectedNetworkMsg string
		closeConnection    bool
		RequiresConn       bool
		expectedError      bool
	}{
		{"Liste Uzunluğu max 2", []string{"emir"}, "Please provide a description that conforms to the specified format.[/whisper <user_name> : <message>]", "", false, false, true},
		{"Listeyi ':' 2 Parça Böl", []string{"emir", "merhaba"}, "Please put a colon (:) between the username and the message.", "", false, false, true},
		{"Mesaj Doğru Bağlantı Hatası", []string{"emir", ":", "merhaba"}, "Error: You are not connected to any server to whisper!", "", false, false, true},
		{"Bağlantı Var Ama Kopuyor", []string{"emir", ":", "merhaba"}, "Failed to send whisper.", "", true, true, true},
		{"Her şey Başarılı", []string{"emir", ":", "merhaba"}, "[whisper to emir]: merhaba", "WHISPER|emir|merhaba\n", false, true, false},
	}

	for _, test := range tests {
		if test.RequiresConn == true {

			whisperChan := make(chan string, 1)
			clientConn, serverConn := net.Pipe()
			hand := &Handler{ServerConn: serverConn}

			if test.closeConnection == true {
				serverConn.Close()
			} else {
				go func() {
					buf := make([]byte, 1024)

					n, readErr := clientConn.Read(buf)
					if readErr == nil {

						actualNetworkMsg := string(buf[:n])
						if actualNetworkMsg != test.expectedNetworkMsg {
							t.Errorf("Network error -> Expected:%q , Received:%q", test.expectedNetworkMsg, actualNetworkMsg)
						}
					}
				}()
			}

			whisperErr := hand.CommandWhisper(whisperChan, test.sent...)
			hasError := whisperErr != nil
			if hasError != test.expectedError {
				t.Fatalf("%s : Excpected:%t , Received:%q", test.testName, test.expectedError, whisperErr)
			}

			take := <-whisperChan
			if take != test.expected {
				t.Errorf("Expected:%q , Received:%q", test.expected, take)
			}

		} else {
			whisperChan := make(chan string, 1)
			hand := &Handler{}

			whisperErr := hand.CommandWhisper(whisperChan, test.sent...)
			hasError := whisperErr != nil
			if hasError != test.expectedError {
				t.Fatalf("%s : Excpected:%t , Received:%q", test.testName, test.expectedError, whisperErr)
			}

			take := <-whisperChan
			if take != test.expected {
				t.Errorf("Expected:%q , Received:%q", test.expected, take)
			}
		}
	}
}
