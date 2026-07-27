package command

import (
	"net"
	"testing"
)

func TestCommandStatus(t *testing.T) {

	tests := []struct {
		TestName           string
		Sent               []string
		Expected           string
		ExpectedNetworkMsg string
		CloseConnection    bool
		RequiresConn       bool
	}{
		{"Belirtilmeyen Durum", []string{}, "SYSTEM| Please specify a status. Usage: /status <online|away|dnd>", "", false, false},
		{"Geçersiz Argüman", []string{"afk"}, "SYSTEM| Invalid status! Use: online, away, dnd", "", false, false},
		{"Sunucu Bağlantısı Yok", []string{"away"}, "SYSTEM| Error: You are not connected to the server.", "", false, false},
		{"Bağlantı var Veri Hatalı", []string{"dnd"}, "SYSTEM| Error sending status to server.", "STATUS|dnd\n", true, true},
		{"Başarılı Durum", []string{"away"}, "SYSTEM| Status change request sent: away", "STATUS|away\n", false, true},
	}

	for _, test := range tests {

		if test.RequiresConn == true {
			statusChan := make(chan string, 1) // karar !
			clientConn, serverConn := net.Pipe()
			hand := &Handler{ServerConn: clientConn}

			if test.CloseConnection == true {
				serverConn.Close()
			} else {
				go func() {
					buf := make([]byte, 1024)

					n, readErr := serverConn.Read([]byte(buf))
					if readErr == nil {

						actualNetworkMsg := string(buf[:n])
						if actualNetworkMsg != test.ExpectedNetworkMsg {
							t.Errorf("Network error -> Expected:%q , Received:%q", test.ExpectedNetworkMsg, actualNetworkMsg)
						}
					}
				}()
			}

			statusErr := hand.CommandStatus(statusChan, test.Sent...)
			if statusErr != nil {
				t.Fatal("CommandStatus returned an unexpected error.")
			}

			take := <-statusChan
			if test.Expected != take {
				t.Errorf("Expected:%q,Received:%q", test.Expected, take)
			}

		} else {

			statusChan := make(chan string, 1) // karar !
			hand := &Handler{}

			statusErr := hand.CommandStatus(statusChan, test.Sent...)
			if statusErr != nil {
				t.Fatal("CommandStatus returned an unexpected error.")
			}

			take := <-statusChan
			if test.Expected != take {
				t.Errorf("Expected:%q,Received:%q", test.Expected, take)
			}
		}
	}
}
