# 🕵️‍♂️ Go-Chatter: P2P Secret Terminal Chat

Go-Chatter is a decentralized, blazing-fast, and completely secure CLI-based chat application built with Go. It utilizes a Peer-to-Peer (P2P) architecture with dynamic leader election, concurrent message broadcasting, and a "Secret Room" philosophy powered by deterministic port hashing.

## ✨ Key Features

* **Decentralized Architecture (P2P):** No central database. The first user to open the application becomes the "Leader" node automatically. If the leader disconnects, the system instantly runs a re-election to keep the network alive.
* **Secret Rooms (Hash-Based Ports):** Rooms are completely hidden. The application generates a deterministic TCP port using the `hash/fnv` algorithm based on the room name. You can only find a room if you know its exact name.
* **Password Protection (The Bouncer):** Secure your private rooms with passwords. The system uses a custom TCP Handshake mechanism to verify credentials before granting access to the room's message channel.
* **High-Performance Concurrency:** Handles hundreds of users simultaneously without breaking a sweat, thanks to Go's native `Goroutines`, `Channels`, and `Mutex` locks preventing race conditions.
* **Command Pattern Design:** Clean, scalable, and easy-to-maintain codebase structure for terminal commands.

## 🚀 Getting Started

### Prerequisites
* [Go](https://golang.org/doc/install) (1.18 or higher)

### Installation
Clone the repository and run the application:

```bash
git clone [https://github.com/Racemir/go-chatter.git](https://github.com/Racemir/go-chatter.git)
cd go-chatter
go run cmd/main.go