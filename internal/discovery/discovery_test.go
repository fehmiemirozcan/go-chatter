package discovery

import (
	"net"
	"testing"
	"time"
)

func TestDiscover_LeaderIp(t *testing.T) {

	// Ağda çalışan bir dinleyici var
	go StartUDPListener("6666")

	time.Sleep(time.Millisecond * 350)

	answer, discoErr := DiscoverLeaderIP("6666")
	if discoErr != nil {
		t.Error("Unexpected error")
	}
	if answer == "" {
		t.Error("The function returned an empty IP.")
	}
}

func TestDiscover_NoOneOnline(t *testing.T) {

	// Ağda çalışan bir dinleyici yok
	answer, discoErr := DiscoverLeaderIP("8889")

	if discoErr == nil {
		t.Error("There is no one on the network; the function should throw an error.")
	}

	timeout := "Timeout"
	if answer != timeout {
		t.Errorf("Expected:%s, Received:%s", timeout, answer)
	}
}

func TestDiscover_WrongAsnwer(t *testing.T) {

	//Ağda yanlış veri gönderilme durumu
	testPort := "9983"
	go func() {
		addr, _ := net.ResolveUDPAddr("udp", ":"+testPort)
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
		t.Error("No error occurred.")
	}
	unknownAnswer := "unknown_answer"
	if discoErr.Error() != unknownAnswer {
		t.Errorf("Expeceted:%s, Received:%v", unknownAnswer, discoErr.Error())
	}

}
