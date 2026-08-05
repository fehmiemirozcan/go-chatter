package command

import (
	"net"
	"testing"
)

func TestCommandNewroom(t *testing.T) {
	// listenin uzunluğu 1 den küçük ise
	// listenin uzunluğu 2 den büyük ise
	// Bağlantı var ama kopuyor
	// her şey başarılı

	tests := []struct {
		testName           string
		sent               []string
		expected           string // kanaldan bekledğimi mesaj
		expectedNetworkMsg string // servera göndereceğim mesaj
		expectedError      bool   // fonksiyon hata msj için
		requiresConn       bool   // conn gereksinimi
		closeConnection    bool
	}{
		{"liste 1'den Küçük", []string{}, "Please specify a room name. Example: /join <room_name> <password>", "", true, false, false},
		{"liste 2'den büyük", []string{"denemeOdası", "123", "merhaba"}, "Error: Too many arguments. Example: /join <room_name> [password]", "", true, false, false},
		{"Bağlantı yok", []string{"denemeOdası", "123"}, "Error: You are not connected to any server to action!", "", true, false, false},
		{"Bağlantı Var Ama Kopuyor", []string{"denemeOdası", "123"}, "Error: Could not send command to server.", "", true, true, true},
		{"Herşey Başarılı", []string{"denemeOdası", "123"}, "", "CMD|join|denemeOdası|123\n", false, true, false},
	}

	for _, test := range tests {
		chanJoin := make(chan string, 1)
		if test.requiresConn == true {

			clientConn, serverConn := net.Pipe()
			hand := &Handler{
				ServerConn:    serverConn,
				RoomPasswords: make(map[string]string)}

			if test.closeConnection == true {
				serverConn.Close()

			} else {
				go func() {
					buf := make([]byte, 1024)
					n, readErr := clientConn.Read(buf)

					if readErr == nil {
						actualNetwork := string(buf[:n])
						if actualNetwork != test.expectedNetworkMsg {
							t.Errorf("Network error ->> Expected:%q , Received:%q", test.expectedNetworkMsg, actualNetwork)
						}
					}
				}()
			}

			joinErr := hand.CommandJoinRoom(chanJoin, test.sent...)
			hasError := joinErr != nil
			if hasError != test.expectedError {
				t.Fatalf("%s Expected:%t , Received:%t", test.testName, test.expectedError, hasError)
			}

			if test.expected != "" {
				take := <-chanJoin
				if test.expected != take {
					t.Errorf("[%s] Output Error ->> Expected:%q , Received:%q", test.testName, test.expected, take)
				}
			}

		} else {

			hand := &Handler{RoomPasswords: make(map[string]string)}

			joinErr := hand.CommandJoinRoom(chanJoin, test.sent...)
			hasError := joinErr != nil
			if hasError != test.expectedError {
				t.Fatalf("[%s] Expected:%t , Received:%t", test.testName, test.expectedError, hasError)
			}

			if test.expected != "" {
				take := <-chanJoin
				if test.expected != take {
					t.Errorf("Expected:%q , Received:%q", test.expected, take)
				}
			}
		}
	}

	// CREATEROOM KISMI
	expectedNetworkMsg := "CMD|createroom|denemeOdası|123\n"
	sent := []string{"denemeOdası", "123"}
	chanCreateroom := make(chan string, 1)
	clientConn, serverConn := net.Pipe()
	hand := &Handler{ServerConn: serverConn, RoomPasswords: make(map[string]string)}

	go func() {
		buf := make([]byte, 1024)
		n, readErr := clientConn.Read(buf)
		if readErr == nil {
			actualNetwork := string(buf[:n])
			if actualNetwork != expectedNetworkMsg {
				t.Errorf("Network Error ->> Expected:%q, Received:%q", expectedNetworkMsg, actualNetwork)
			}
		}
	}()

	createroomErr := hand.CommandCreateRoom(chanCreateroom, sent...)
	hasError := createroomErr != nil
	if hasError != false {
		t.Fatalf("Expected:%t , Received:%t", false, hasError)
	}
}
