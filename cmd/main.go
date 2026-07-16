package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Racemir/go-chatter/internal/client"
	"github.com/Racemir/go-chatter/internal/discovery"
	"github.com/Racemir/go-chatter/internal/server"
	"github.com/Racemir/go-chatter/internal/tui"
)

var (
	inChan  = make(chan string)
	outChan = make(chan string)
)

func main() {
	fmt.Println("")
	fmt.Print(">>> Welcome Go Chatter <<<\n")

	userName := getUserName()

	fmt.Println("Network is being scanned, searching for a Leader...")
	// Lider arayışı kapsamında ağ taranıyor.

	leaderIP, discoveryErr := discovery.DiscoverLeaderIP()
	if discoveryErr != nil {
		fmt.Println("No leader found. YOU are the Leader!")
		// Lider bulunamadı. Lider SİZSİNİZ!
		startAsLeader(userName)
	} else {
		fmt.Printf("Leader found! IP Address: %s", leaderIP)
		// Lider bulundu! IP Adresi:
		startAsFollower(userName, leaderIP)
	}

}

func getUserName() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		fmt.Println("Name cannot be empty. Default: Guest.")
		// İsim boş bırakılamaz. Varsayılan: Misafir.
		return "Guest"
	}
	return name
}

func startAsLeader(userName string) {
	// Bulunabilir olsun diye dinleyiciyi açtık.
	go discovery.StartUDPListener()

	// Sunucuyu ayağa kaldırıyoruz.
	hub := server.NewRouterHub()
	go hub.StartServer(":9000")

	time.Sleep(100 * time.Millisecond)

	// Kurduğumuz sunucuya bağlanıyoruz.
	conn, dialErr := net.Dial("tcp", "localhost:9000")
	if dialErr != nil {
		fmt.Println("Failed to start Master Server!")
		return
	}
	// Global odaya Handshake
	conn.Write([]byte(fmt.Sprintf("createroom|global||%s", userName)))

	// Sunucudan onay (Successful) bekliyoruz.
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	if string(buf[:n]) == "Successful" {
		go client.StartChat(userName, conn, inChan, outChan)
		tui.StartUI(inChan, outChan, userName)
	} else {
		fmt.Println("Server Error:", string(buf[:n]))
		os.Exit(1) // Onay gelmediyse uygulamayı kapat
	}
}

func startAsFollower(userName, leaderIP string) {

	address := fmt.Sprintf("%s:9000", leaderIP)
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		fmt.Println("Failed to connect to the Leader!")
		// Liderle bağlantı kurulamadı!
		return
	}

	// Global odaya giriş izni Handshake
	conn.Write([]byte(fmt.Sprintf("join|global||%s", userName)))

	// Liderin sunucusundan onay bekliyoruz.
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	if string(buf[:n]) == "Successful" {
		go client.StartChat(userName, conn, inChan, outChan)
		tui.StartUI(inChan, outChan, userName)
	} else {
		fmt.Println("\n[!] Connection rejected:", string(buf[:n]))
		os.Exit(1)
		// Bağlantı reddedildi.
	}

}
