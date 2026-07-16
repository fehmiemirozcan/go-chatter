package client

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Racemir/go-chatter/internal/command"
	"github.com/Racemir/go-chatter/internal/discovery"
	"github.com/Racemir/go-chatter/internal/security"
	"github.com/Racemir/go-chatter/internal/server"
)

type ChatState struct {
	conn net.Conn
}

func StartChat(userName string, globalConn net.Conn, inputChan chan string, outputChan chan string) error {

	cmdHandler := &command.Handler{
		UserName:      userName,
		RoomPasswords: make(map[string]string),
		CurrentRoom:   "global",
		ServerConn:    globalConn,
	}

	cmdHandler.RoomPasswords["global"] = ""

	commanddictionary := cmdHandler.GetCommands()

	// BroadcastRooms komutu command.go'dan silindiği için buradaki çağrıyı kaldırdık.

	outputChan <- "Welcome to the secret chat"
	cmdHandler.CommandHelp(outputChan)

	//kalavyeden gelen msj ilet
	deathSignal := make(chan struct{})

	var listenToServer func(roomName string, c net.Conn)
	listenToServer = func(roomName string, c net.Conn) {
		readBuf := make([]byte, 1024)
		for {
			n, readErr := c.Read(readBuf) // Sunucudan mesaj var mı ?
			if readErr != nil {
				time.Sleep(50 * time.Second)
				if cmdHandler.CurrentRoom != roomName {
					return
				} else {
					msg := fmt.Sprintf("\n[!] Leader connection lost! Initial host migration: %s\n", roomName)
					outputChan <- msg
					// oda benim isteğim dışında bağlantısını kopardu hala aynı odadayım

					// Ağda yeni bir lider var mı?
					newLeaderIP, leaderErr := discovery.DiscoverLeaderIP()

					// Lider Yoksa
					if leaderErr != nil {
						outputChan <- "[!] You are the leader! Starting server..."

						// UDP dinleyicisini ve ana sunucumuzu ayağa kaldırıyoruz
						go discovery.StartUDPListener()
						hub := server.NewRouterHub()
						go hub.StartServer(":9000")

						time.Sleep(100 * time.Millisecond)

						// TCP ile bağlanıyoruz
						newConn, _ := net.Dial("tcp", "localhost:9000")

						msgs := fmt.Sprintf("createroom|%s||%s", roomName, cmdHandler.UserName)
						newConn.Write([]byte(msgs))

						buf := make([]byte, 1024)
						n, _ := newConn.Read(buf)

						if string(buf[:n]) == "Successful" {
							cmdHandler.ServerConn = newConn
							go listenToServer(roomName, newConn)
						}

					} else {
						// Lider Varsa
						msg := fmt.Sprintf("[!] New leader found at %s. Connecting...", newLeaderIP)
						outputChan <- msg

						localhost := fmt.Sprintf("%s:9000", newLeaderIP)
						newConn, _ := net.Dial("tcp", localhost)

						msgs := fmt.Sprintf("join|%s||%s", roomName, cmdHandler.UserName)
						newConn.Write([]byte(msgs))

						// Onay alırsak yeni bağlantıyı kaydet ve dinlemeye başla
						buf := make([]byte, 1024)
						n, _ := newConn.Read(buf)
						if string(buf[:n]) == "Successful" {
							cmdHandler.ServerConn = newConn
							go listenToServer(roomName, newConn)
						} else {
							outputChan <- "[!] Migration failed: Connection rejected by new leader."
						}
					}
				}

				return // oda değiştirdim sessizce diğer odadan ayrıl
			}

			rawMsg := string(readBuf[:n])

			// TCP'nin birleştirdiği paketleri \n'den ayırıyoruz
			message := strings.Split(rawMsg, "\n")

			for _, msg := range message {
				msg = strings.TrimSpace(msg)

				if msg == "" {
					continue
				}

				// Oda Değişikliği Başarılı
				if strings.HasPrefix(msg, "ROOMCHANGE|SUCCESS|") {
					newRoom := strings.TrimPrefix(msg, "ROOMCHANGE|SUCCESS|")
					cmdHandler.CurrentRoom = newRoom
					outputChan <- fmt.Sprintf("[SYSTEM] Successfully moved to room '%s'!\n", newRoom)
					continue
				}

				// Oda Değişikliği Hatalı
				if strings.HasPrefix(msg, "ROOMCHANGE|ERR|") {
					errMsg := strings.TrimPrefix(msg, "ROOMCHANGE|ERR|")
					outputChan <- fmt.Sprintf("[ERROR] Room action failed: %s\n", errMsg)
					continue
				}

				// Oda Listesi
				if strings.HasPrefix(msg, "ALLROOMS|") {
					roomList := strings.TrimPrefix(msg, "ALLROOMS|")
					formatTui := fmt.Sprintf("ROOMS|%s|%s", roomList, cmdHandler.CurrentRoom)
					outputChan <- formatTui
					continue
				}

				// Kullanıcı Listesi
				if strings.HasPrefix(msg, "ROOMUSERS|") {
					userList := strings.TrimPrefix(msg, "ROOMUSERS|")
					formatTui := fmt.Sprintf("USERS|%s", userList)
					outputChan <- formatTui
					continue
				}

				if strings.HasPrefix(msg, "WHISPER|") {
					// "WHISPER|%s|%s\n
					args := strings.SplitN(msg, "|", 3)
					if len(args) == 3 {
						senderUser := args[1]
						message := args[2]
						formatTui := fmt.Sprintf("WHISPER|%s|%s", senderUser, message)
						outputChan <- formatTui
					}
					continue
				}

				// Normal mesajlar (Artık odaya göre kısıtlamıyoruz, tek bağlantı var)
				incomingStr := msg
				parts := strings.SplitN(incomingStr, ": ", 2)

				if len(parts) == 2 {
					senderName := parts[0]
					encryptedMsg := strings.TrimSpace(parts[1]) // Şifrelenmiş msg

					currentPass := cmdHandler.RoomPasswords[cmdHandler.CurrentRoom] // Şifre
					decryptedMsg := security.DecryptMessage(encryptedMsg, currentPass)

					formatMsg := fmt.Sprintf("MSG| %s: %s\n", senderName, decryptedMsg)
					outputChan <- formatMsg
				} else {
					outputChan <- strings.TrimSpace(incomingStr)
					// İçinde : yoksa sistem msg
				}
			}

		}
	}
	// Global oda her zaman ilk dinlenecek.
	go listenToServer("global", globalConn)
	var lastClientMsgTime time.Time

	// Klavyeden mesaj okuma.
	for {
		select {
		case <-deathSignal:
			outputChan <- "Connection completely lost"
			// Bağlantı tamamen kesildi
			return fmt.Errorf("Connection completely lost")
		case rawInput := <-inputChan:

			// Gelen metnin başındaki/sonundaki boşlukları temizliyorum.
			rawInput = strings.TrimSpace(rawInput)

			// Boşsa hiç bir şey yaptırmıyoruz.
			if rawInput == "" {
				continue
			}

			if strings.HasPrefix(rawInput, "/") {
				// Gelen texti komut ve msg olarak 2'ye ayır.
				words := cleanInput(rawInput)

				commandName := words[0]
				args := []string{}
				if len(words) > 1 {
					args = words[1:]
				}

				// GetCommand fonksiyonu çağır.
				cmd, ok := commanddictionary[commandName]
				if ok {
					// Eğer komut /quit ise
					if commandName == "/quit" {
						if cmdHandler.ServerConn != nil {
							cmdHandler.ServerConn.Close() // Fişi çek!
						}
					}
					cmd.Callback(outputChan, args...)

				} else {
					outputChan <- "Unknown command. Type /help"
				}

			} else {
				// Normal Mesaj Gönderme Bloğu
				now := time.Since(lastClientMsgTime)
				if now < 300*time.Millisecond {
					// Ağa hiç göndermeden direkt kendi ekranına uyarı bas ve işlemi kes!
					outputChan <- "SYSTEM| You're sending messages too quickly!"
					continue
				}
				lastClientMsgTime = time.Now()

				if cmdHandler.ServerConn == nil {
					outputChan <- "Error: You are not connected to the server."
					continue
				}

				currentPass := cmdHandler.RoomPasswords[cmdHandler.CurrentRoom] // Mevcut şifre
				encryptedText := security.EncryptMessage(rawInput, currentPass) // kilitle

				msg := encryptedText + "\n"
				_, writeErr := cmdHandler.ServerConn.Write([]byte(msg))
				if writeErr != nil {
					outputChan <- "Message could not be sent !!"
					// Mesaj gönderilemedi!!
				} else {
					myMsg := fmt.Sprintf("ME| %s", rawInput)
					outputChan <- myMsg
				}
			}
		}
	}
}

// cleanInput fonksiyonu.
func cleanInput(dirtWords string) []string {
	output := strings.ToLower(dirtWords)

	// dirtWords := "   /join      GizliOda   "
	// ["/join", "gizlioda"]
	words := strings.Fields(output)

	return words
}
