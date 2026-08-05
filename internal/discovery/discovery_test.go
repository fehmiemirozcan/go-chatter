package discovery

import (
	"net"
	"testing"
	"time"
)

func TestStartUDPListener_Error(t *testing.T) {
	testPort := "7777"

	addr, _ := net.ResolveUDPAddr("udp", ":"+testPort)
	dummyConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Test setup failed: could not occupy port: %v", err)
	}

	defer dummyConn.Close()

	// Fonksiyonunu aynı port ile çağırdım port zaten dummyConn tarafından dolu olduğu için, içerideki ListenUDP
	// "address already in use" hatası verecek ve direkt return edecektir.
	StartUDPListener(testPort)
}
func TestDiscover_ResolUDPAddrEror(t *testing.T) {
	testPort := "-1"

	answer, discoErr := DiscoverLeaderIP(testPort)
	if discoErr == nil {
		t.Fatalf("Expected a resolution error due to invalid port, but got no error")
	}

	if answer != "" {
		t.Errorf("Expected empty string on error, but received: %s", answer)
	}
}
func TestDiscover_LeaderIp(t *testing.T) {
	testPort := "6666"
	// Ağda çalışan bir dinleyici var
	go StartUDPListener(testPort)

	time.Sleep(time.Millisecond * 100)

	answer, discoErr := DiscoverLeaderIP(testPort)
	if discoErr != nil {
		t.Fatalf("Unexpected error: %v", discoErr)
	}
	if answer == "" || answer == "Timeout" {
		t.Errorf("The function returned an empty IP or Timeout. Received: %s", answer)
	}
}

func TestDiscover_NoOneOnline(t *testing.T) {
	testPort := "8889"
	// Ağda çalışan bir dinleyici yok
	answer, discoErr := DiscoverLeaderIP(testPort)

	if discoErr == nil {
		t.Fatal("There is no one on the network; the function should throw an error.")
	}

	expectedAnswer := "Timeout"
	if answer != expectedAnswer {
		t.Errorf("Expected: %s, Received: %s", expectedAnswer, answer)
	}
}

func TestDiscover_WrongAsnwer(t *testing.T) {
	testPort := "9983"
	//Ağda yanlış veri gönderilme durumu

	go func() {
		addr, resolErr := net.ResolveUDPAddr("udp", ":"+testPort)
		if resolErr != nil {
			return
		}
		conn, listenErr := net.ListenUDP("udp", addr)
		if listenErr != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, clientAddr, _ := conn.ReadFromUDP(buf)

		if string(buf[:n]) == "who_is_the_leader" {
			conn.WriteToUDP([]byte("Uzaydan gelen yeni lider benim"), clientAddr)
		}
	}()

	time.Sleep(time.Millisecond * 100)

	_, discoErr := DiscoverLeaderIP(testPort)
	if discoErr == nil {
		t.Fatal("An error was expected because of the wrong answer, but no error occurred.")
	}
	expectedErrMessage := "unknown_answer"
	if discoErr.Error() != expectedErrMessage {
		t.Errorf("Expected Error: %s, Received Error: %v", expectedErrMessage, discoErr.Error())
	}
}
