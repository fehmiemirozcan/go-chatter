package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/Racemir/goserver/go-chatter/internal"
)

func main() {

	fmt.Print(">>> Welcome Go Chatter <<<\n")

	_, err := internal.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	conn, err := net.Dial("tcp", "yqywc-159-146-3-206.a.free.pinggy.link:42467")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go func() {
		io.Copy(os.Stdout, conn)
	}()
	for {
		fmt.Print("> ")
		msg := internal.GetInput()

		fmt.Fprintf(conn, msg+"\n")
	}
}
