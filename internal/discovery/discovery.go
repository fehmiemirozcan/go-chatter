package discovery

import (
	"fmt"
	"net"
	"time"
)

func DiscoverLeaderIP(port string) (string, error) {

	// Tüm cihazlara kısmı
	broadcastAddr, resolUDPAddrErr := net.ResolveUDPAddr("udp", "127.0.0.1:"+port)
	if resolUDPAddrErr != nil {
		return "", resolUDPAddrErr
	}

	conn, dialUDPErr := net.DialUDP("udp", nil, broadcastAddr)
	if dialUDPErr != nil {
		return "", dialUDPErr
	}

	defer conn.Close()

	_, writeErr := conn.Write([]byte("who_is_the_leader")) // lider kim ?
	if writeErr != nil {
		return "", writeErr
	}

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	// lider var mı varsa kim cevabı gelene kadar bekletiyoruz.

	buf := make([]byte, 1024)
	n, serverAddr, readErr := conn.ReadFromUDP(buf)
	if readErr != nil {
		return "Timeout", readErr
	}

	if string(buf[:n]) == "I_am_the_leader" { // lider benim.
		return serverAddr.IP.String(), nil
	}

	return "", fmt.Errorf("unknown_answer")
	// Bilinmeyen_yanit
}

func StartUDPListener(port string) {

	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, listenUDPErr := net.ListenUDP("udp", addr)
	if listenUDPErr != nil {
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		// lider_kim
		n, clientAddr, readFromErr := conn.ReadFromUDP(buf)
		if readFromErr != nil {
			continue
		}

		if string(buf[:n]) == "who_is_the_leader" {
			conn.WriteToUDP([]byte("I_am_the_leader"), clientAddr)
		}
	}
}
