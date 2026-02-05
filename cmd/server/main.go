package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

var clientList = make(map[net.Conn]bool)
var mutex sync.Mutex

func main() {

	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Print("Welcome to the server!\n")
	for {
		con, err := conn.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			mutex.Lock()
			clientList[c] = true
			mutex.Unlock()
			defer c.Close()

			defer func() {
				mutex.Lock()
				delete(clientList, c)
				fmt.Printf("%v user left", c)
				mutex.Unlock()
			}()

			scanner := bufio.NewScanner(c)
			for scanner.Scan() {
				msg := scanner.Text()
				msg = msg + "\n"
				mutex.Lock()
				for a := range clientList {
					if a != c {
						fmt.Fprintf(a, msg)
						fmt.Fprintf(a, "> ")
					}
				}
				mutex.Unlock()
			}
		}(con)
	}
}
