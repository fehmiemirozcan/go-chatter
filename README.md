<div align="center">

# 🕵️‍♂️ Go-Chatter

### Ephemeral, Decentralized P2P Terminal Chat with End-to-End Room Encryption

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Security](https://img.shields.io/badge/Encryption-AES--256--GCM-success?style=flat)](https://en.wikipedia.org/wiki/Galois/Counter_Mode)
[![Persistence](https://img.shields.io/badge/Logs-Zero%20Persistence%20(RAM%20Only)-red?style=flat)](#-privacy--zero-persistence-philosophy)
[![TUI](https://img.shields.io/badge/TUI-Charmbracelet%20Bubble%20Tea-f43f5e?style=flat)](https://github.com/charmbracelet/bubbletea)
[![Architecture](https://img.shields.io/badge/Network-P2P%20%2F%20Host%20Migration-blueviolet?style=flat)](#-p2p-architecture--host-migration)

[ **English** ](README.md) | [ **Türkçe** ](README.tr.md)

</div>

---

## 🔒 Privacy & Zero-Persistence Philosophy

**Go-Chatter** is built from the ground up for strictly private, ephemeral communication inside your terminal. It operates under a **zero-trace, zero-persistence** policy:

* **No Database:** No SQL, NoSQL, or external datastore is used.
* **No Disk Logs:** Messages, usernames, and room passwords are never written to disk or recorded in log files.
* **RAM-Only Lifecycle:** All communications flow exclusively through volatile memory channels and active network buffers.
* **Instant Purge on Room Switch:** Whenever you switch rooms, previous chat buffers are immediately cleared from memory.
* **Zero History on Exit:** The moment you close the application or terminate the connection, all session data disappears completely.

---

## ✨ Features

- **🌐 Global & Private Rooms:** Jump right into the public `global` room or create password-protected secret rooms with `/createroom`.
- **🔐 AES-256-GCM Authenticated Encryption:**
  - Password-protected rooms automatically encrypt all outgoing messages using **AES-256-GCM**.
  - 32-byte cryptographic keys are derived via **SHA-256** from room passwords.
  - Every message includes a cryptographically secure 12-byte random nonce (`crypto/rand`) and a 16-byte MAC authentication tag to prevent tampering or replay attacks.
- **🔄 P2P Auto-Discovery & Dynamic Leader Election:**
  - No dedicated central server required.
  - The first node on the local network automatically elects itself as the **Leader** and hosts the TCP router hub.
  - Subsequent nodes automatically discover the leader via **UDP broadcast**.
- **⚡ Fast Host Migration:**
  - If the leader disconnects or drops offline, connected nodes automatically detect the loss and execute a seamless leader re-election to keep the network alive without manual intervention.
- **💬 Direct Whisper (1-on-1 Messages):**
  - Send private messages directly to any user across any room using `/whisper <username> : <message>`.
- **🟢 Presence & Status Management:**
  - Update your status in real time (`online`, `away`, `dnd`) using `/status`. Status updates are instantly reflected in the room users list.
- **🛡️ Abuse & Spam Protections:**
  - **Rate Limiting:** 300ms cooldown timer on messages enforced on both client and server to prevent spamming.
  - **Connection Limiting:** Maximum 3 concurrent connections allowed per IP address to safeguard network resources.
  - **Nickname Protection:** Case-insensitive, Unicode-normalized duplicate name prevention.
- **🎨 Modern Terminal UI (TUI):**
  - Built with **Charmbracelet Bubble Tea**, **Lipgloss**, and **Bubbles**.
  - Multi-pane responsive layout: Chat stream, Rooms explorer, and Room users list.
  - Complete **Mouse** and **Keyboard** navigation support.
  - Command autocompletion and input history buffer.

---

## 🏗️ Architecture & Network Flow

```
                      +-----------------------------+
                      |   Node Starts Application   |
                      +-----------------------------+
                                     |
                         [UDP Broadcast on :8081]
                        "who_is_the_leader" query
                                     |
                   +-----------------+-----------------+
                   |                                   |
         (No Response / Timeout)             ("I_am_the_leader" received)
                   |                                   |
                   v                                   v
        +----------------------+            +----------------------+
        |      BECOME LEADER   |            |    BECOME FOLLOWER   |
        | - Open UDP :8081     |            | - Connect to Leader  |
        | - Start TCP :9000    |            |   TCP at :9000       |
        | - Host Router Hub    |            | - Join "global" room |
        +----------------------+            +----------------------+
                   ^                                   |
                   |      [Leader Connection Lost]     |
                   +-----------------------------------+
                         Trigger Host Migration
```

### Communication Details:
- **UDP Discovery Port:** `8081` (Leader query & election)
- **TCP Router Port:** `9000` (Message routing, room management, whisper dispatching)

---

## ⌨️ Command Reference

| Command | Usage | Description |
| :--- | :--- | :--- |
| `/help` | `/help` | Displays available commands and usage hints. |
| `/createroom` | `/createroom <room_name> [password]` | Creates a new room (max 3 custom rooms per user). Optional password enables AES-256-GCM encryption. |
| `/join` | `/join <room_name> [password]` | Joins an existing room. Prompts for password if locked. |
| `/whisper` | `/whisper <user_name> : <message>` | Sends a private 1-on-1 message to a specific user. |
| `/status` | `/status <online\|away\|dnd>` | Updates user presence status (`online`, `away`, `dnd`). |
| `/quit` | `/quit` | Disconnects and terminates the application cleanly. |

---

## 🎮 TUI Navigation & Shortcuts

| Key / Mouse Action | Action |
| :--- | :--- |
| `Tab` | Switch focus between **Input** ➔ **Chat Viewport** ➔ **Rooms Panel** ➔ **Users Panel**.<br>*(If input starts with `/`, cycles through command autocompletion).* |
| `Esc` | Clears active search filter or returns focus immediately to the message input bar. |
| `Up` / `Down` | In Input bar: Browse through sent message history.<br>In Lists / Chat: Scroll items. |
| `Enter` | Send message, join selected room, or initiate whisper to selected user. |
| `/` (in Rooms/Users) | Trigger fast interactive search / filter in the focused list. |
| `Mouse Left Click` | Focus any pane or select rooms / users directly. Clicking a locked room prepares `/join`, clicking a user prepares `/whisper`. |
| `Mouse Scroll Wheel`| Scroll through chat history or room and user lists. |
| `Ctrl + C` | Force quit the application. |

---

## 🛠️ Tech Stack & Dependencies

Strictly built using modern Go and the Charmbracelet terminal ecosystem:

* **Language:** [Go](https://go.dev/) (`1.25+` / standard library)
  - `net`: Raw TCP router and UDP discovery implementation.
  - `crypto/aes`, `crypto/cipher`, `crypto/sha256`, `crypto/rand`: Authenticated AES-256-GCM encryption.
  - `sync`: Goroutine synchronization and mutex locks for thread-safe state management.
* **Terminal User Interface:**
  - `github.com/charmbracelet/bubbletea`: Functional Elm-inspired TUI runtime.
  - `github.com/charmbracelet/lipgloss`: Style definitions, colors, and layout engine.
  - `github.com/charmbracelet/bubbles`: Textinput, Viewport, and List UI components.

---

## 🚀 Getting Started

### Prerequisites
* [Go](https://go.dev/dl/) installed on your machine (`1.21` or higher recommended, built with `1.25.1`).

### Installation & Run

1. **Clone the repository:**
   ```bash
   git clone https://github.com/fehmiemirozcan/go-chatter.git
   cd go-chatter
   ```

2. **Run the application:**
   ```bash
   go run ./cmd/main.go
   ```
   *Alternatively, using the Makefile:*
   ```bash
   make run
   ```

3. **Join with Multiple Users:**
   - Launch another terminal instance on the same machine or local network:
     ```bash
     go run ./cmd/main.go
     ```
   - Enter your nickname and start chatting immediately.

---

## 🧪 Testing

Run the automated test suite covering client, commands, discovery, security, server hub, and TUI modules:

```bash
# Run all unit and integration tests
make test
# Or directly with Go:
go test ./...
```

To generate a test coverage report:
```bash
make coverage
```

---

## 📁 Project Structure

```
go-chatter/
├── cmd/
│   └── main.go                  # Application entry point, leader check & node bootstrap
├── internal/
│   ├── client/
│   │   ├── chat.go              # Client message loops, incoming parser & migration logic
│   │   └── chat_test.go         # Client unit & migration tests
│   ├── command/
│   │   ├── command_exit.go      # /quit command implementation
│   │   ├── command_get.go       # Command dictionary & handler definition
│   │   ├── command_help.go      # /help command implementation
│   │   ├── command_roomactual.go# /createroom & /join room operations
│   │   ├── command_status.go    # /status presence handler
│   │   ├── command_whisper.go   # /whisper direct messaging handler
│   │   └── *_test.go            # Command suite tests
│   ├── discovery/
│   │   ├── discovery.go         # UDP broadcast search & leader listener (:8081)
│   │   └── discovery_test.go    # Discovery tests
│   ├── security/
│   │   ├── security.go          # SHA-256 key generation, AES-256-GCM Encrypt/Decrypt
│   │   └── security_test.go     # Encryption verification tests
│   ├── server/
│   │   ├── server.go            # RouterHub, TCP listener (:9000), room & user management
│   │   └── server_test.go       # Server concurrency & protocol tests
│   └── tui/
│       ├── styles.go            # Lipgloss styling tokens and colors
│       ├── tui.go               # Bubble Tea program initialization
│       ├── types.go             # TUI data models and message definitions
│       ├── update.go            # Key/mouse events, focus management & commands
│       ├── view.go              # UI layout renderer (Chat, Rooms, Users panels)
│       └── listenMessage_test.go# TUI message parsing tests
├── Makefile                     # Build, run, and test workflows
├── go.mod                       # Module definition & dependencies
└── go.sum                       # Dependency checksums
```