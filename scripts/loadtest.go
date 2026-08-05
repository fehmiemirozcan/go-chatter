package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	serverAddress = "localhost:9000"
	botCount      = 4
)

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= botCount; i++ {
		wg.Add(1)
		go startBot(fmt.Sprintf("Bot_%d", i), &wg)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func startBot(name string, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("[%s] Server Not Found: %v\n", name, err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("join|global||%s\n", name)))
	time.Sleep(500 * time.Millisecond)

	if name == "Bot_4" {
		return
	}

	if name == "Bot_1" {
		fmt.Println("Bot_1 Spam Testi")
		for i := 0; i < 5; i++ {
			conn.Write([]byte("Spam mesaji!\n"))

		}
	}

	if name == "Bot_2" {
		fmt.Println("Bot_2 Oda Kurma Limiti Testi")
		conn.Write([]byte("CMD|createroom|oda0\n"))
		time.Sleep(400 * time.Millisecond)
		conn.Write([]byte("CMD|createroom|oda1\n"))
		time.Sleep(400 * time.Millisecond)
		conn.Write([]byte("CMD|createroom|oda2\n"))
		time.Sleep(400 * time.Millisecond)
		conn.Write([]byte("CMD|createroom|oda3\n"))
		time.Sleep(400 * time.Millisecond)
	}

	if name == "Bot_3" {
		fmt.Println("Şifreli oda")
		conn.Write([]byte("CMD|createroom|sifrelioda|123\n"))
	}
	time.Sleep(3 * time.Second)
}
