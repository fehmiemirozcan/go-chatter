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

var (
	DiscoverLeaderFunc = discovery.DiscoverLeaderIP
	StartUDPFunc       = discovery.StartUDPListener
	DialFunc           = net.Dial
	StartServerFunc    = func() error {
		hub := server.NewRouterHub()
		return hub.StartServer(":9000")
	}
)

type ChatState struct {
	conn net.Conn
}

func StartChat(userName string, globalConn net.Conn, inputChan chan string, outputChan chan string, deathSignal chan struct{}) error {

	cmdHandler := &command.Handler{
		UserName:      userName,
		RoomPasswords: make(map[string]string),
		CurrentRoom:   "global",
		ServerConn:    globalConn,
	}

	cmdHandler.RoomPasswords["global"] = ""

	outputChan <- "Welcome to the secret chat"
	cmdHandler.CommandHelp(outputChan)

	//kalavyeden gelen msj ilet

	var listenToServer func(roomName string, c net.Conn)
	listenToServer = func(roomName string, c net.Conn) {
		readBuf := make([]byte, 65536) // 1024*64 = 65.536 (64 kb)
		for {
			n, readErr := c.Read(readBuf) // Sunucuyu dinliyoruz göç süreci için lider çıkarsa
			if readErr != nil {
				// Eğer oda geçişi yaptıysa direk çık
				if cmdHandler.CurrentRoom != roomName {
					return
				}

				outputChan <- fmt.Sprintf("\n[!] Leader connection lost! Fast host migration: %s\n", roomName)
				newLeaderIP, leaderErr := DiscoverLeaderFunc("8081") // ağda lider var mı ?

				// Lider yoksa
				if leaderErr != nil || newLeaderIP == "" || newLeaderIP == "Timeout" {
					startErr := StartServerFunc()
					// Başkası lider olduysa
					if startErr != nil {

						outputChan <- "[!] Another client claimed the throne first! Connecting...\n"

						for i := 0; i < 5; i++ {
							// liderin portuna bağlanmak
							newConn, dialErr := DialFunc("tcp", "localhost:9000")
							if dialErr == nil {
								msgs := fmt.Sprintf("join|%s||%s", roomName, cmdHandler.UserName)
								newConn.Write([]byte(msgs))
								buf := make([]byte, 1024)
								n, _ := newConn.Read(buf)

								if string(buf[:n]) == "Successful" {
									cmdHandler.ServerConn = newConn
									go listenToServer(roomName, newConn)
									return
								}
								newConn.Close()
							}
							// Sunucuyu DDOSlamamak için ufak bir bekleme!
							time.Sleep(100 * time.Millisecond)
						}
						outputChan <- "[!] Migration failed: Connection rejected by new leader.\n"

					} else {
						// sunucu ayağa kalktı ve lider bizsek
						outputChan <- "[!] You are the leader! Starting server...\n"
						go StartUDPFunc("8081")

						for i := 0; i < 5; i++ {
							newConn, dialErr := DialFunc("tcp", "localhost:9000")
							if dialErr == nil {
								msgs := fmt.Sprintf("createroom|%s||%s", roomName, cmdHandler.UserName)
								newConn.Write([]byte(msgs))
								buf := make([]byte, 1024)
								n, _ := newConn.Read(buf)

								if string(buf[:n]) == "Successful" {
									cmdHandler.ServerConn = newConn
									go listenToServer(roomName, newConn)
									return
								}
								newConn.Close()
							}
							time.Sleep(50 * time.Millisecond)
						}
					}
				} else {
					// Ağda zaten lider var
					msg := fmt.Sprintf("[!] New leader found at %s. Connecting...\n", newLeaderIP)
					outputChan <- msg

					leaderAddress := fmt.Sprintf("%s:9000", newLeaderIP)
					for i := 0; i < 5; i++ {
						newConn, dialErr := DialFunc("tcp", leaderAddress)
						if dialErr == nil {
							msgs := fmt.Sprintf("join|%s||%s", roomName, cmdHandler.UserName)
							newConn.Write([]byte(msgs))
							buf := make([]byte, 1024)
							n, _ := newConn.Read(buf)

							if string(buf[:n]) == "Successful" {
								cmdHandler.ServerConn = newConn
								go listenToServer(roomName, newConn)
								return
							}
							newConn.Close()
						}
						time.Sleep(100 * time.Millisecond)
					}
					outputChan <- "[!] Migration failed: Could not connect to remote leader.\n"
				}
				return
			}

			rawMsg := string(readBuf[:n])

			message := strings.Split(rawMsg, "\n")

			for _, msg := range message {

				// Normal mesajlar (Artık odaya göre kısıtlamıyoruz, tek bağlantı var)
				parseIncomingMessage(msg, cmdHandler, outputChan)

			}
		}
		// Global oda her zaman ilk dinlenecek.
	}
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
				handleUserCommand(rawInput, cmdHandler, outputChan)

			} else {
				sendNormalMessage(rawInput, cmdHandler, outputChan, &lastClientMsgTime)
			}
		}
	}
}

// CleanInput fonksiyonu.
func cleanInput(dirtWords string) []string {
	words := strings.Fields(dirtWords)
	if len(words) > 0 {
		words[0] = strings.ToLower(words[0])
	}

	return words
}

// Sunucudan gelen ham mesajı analiz eder ve ilgili kanala yönlendirir.
func parseIncomingMessage(msg string, cmdHandler *command.Handler, outputChan chan string) {
	msg = strings.TrimSpace(msg)

	if msg == "" {
		return
	}

	// Oda Değişikliği Başarılı
	if strings.HasPrefix(msg, "ROOMCHANGE|SUCCESS|") {
		newRoom := strings.TrimPrefix(msg, "ROOMCHANGE|SUCCESS|")
		cmdHandler.CurrentRoom = newRoom
		outputChan <- fmt.Sprintf("[SYSTEM] Successfully moved to room '%s'!\n", newRoom)
		return
	}

	// Oda Değişikliği Hatalı
	if strings.HasPrefix(msg, "ROOMCHANGE|ERR|") {
		errMsg := strings.TrimPrefix(msg, "ROOMCHANGE|ERR|")
		outputChan <- fmt.Sprintf("[ERROR] Room action failed: %s\n", errMsg)
		return
	}

	// Oda Listesi
	if strings.HasPrefix(msg, "ALLROOMS|") {
		roomList := strings.TrimPrefix(msg, "ALLROOMS|")
		formatTui := fmt.Sprintf("ROOMS|%s|%s", roomList, cmdHandler.CurrentRoom)
		outputChan <- formatTui
		return
	}

	// Kullanıcı Listesi
	if strings.HasPrefix(msg, "ROOMUSERS|") {
		userList := strings.TrimPrefix(msg, "ROOMUSERS|")
		formatTui := fmt.Sprintf("USERS|%s", userList)
		outputChan <- formatTui
		return
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
		return
	}

	parts := strings.SplitN(msg, ": ", 2)

	if len(parts) == 2 {
		senderName := parts[0]
		encryptedMsg := strings.TrimSpace(parts[1]) // Şifrelenmiş msg

		currentPass := cmdHandler.RoomPasswords[cmdHandler.CurrentRoom] // Şifre
		decryptedMsg := security.DecryptMessage(encryptedMsg, currentPass)

		formatMsg := fmt.Sprintf("MSG| %s: %s\n", senderName, decryptedMsg)
		outputChan <- formatMsg
	} else {
		// İçinde : yoksa sistem msg
		outputChan <- strings.TrimSpace(msg)
	}
}

// Klavyeden "/" ile başlayan girdileri analiz edip çalıştırır.
func handleUserCommand(rawInput string, cmdHandler *command.Handler, outputChan chan string) {
	// Gelen texti komut ve argümanlar olarak 2'ye ayır
	words := cleanInput(rawInput)
	if len(words) == 0 {
		return
	}

	commandName := words[0]
	args := []string{}
	if len(words) > 1 {
		args = words[1:]
	}
	// Komut sözlüğünü çek ve komutu ara
	commanddictionary := cmdHandler.GetCommands()
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
}

// Kullanıcının yazdığı normal metinleri şifreler, spam kontrolünden geçirir ve sunucuya gönderir.
func sendNormalMessage(rawInput string, cmdHandler *command.Handler, outputChan chan string, lastMsgTime *time.Time) {
	// Spam (Rate Limit) Kontrolü
	now := time.Since(*lastMsgTime)
	if now < 300*time.Millisecond {
		// Ağa hiç göndermeden direkt kendi ekranına uyarı bas ve işlemi kes!
		outputChan <- "SYSTEM| You're sending messages too quickly!"
		return
	}
	*lastMsgTime = time.Now() // Orijinal zamanlayıcıyı güncelliyoruz
	// Bağlantı Kontrolü
	if cmdHandler.ServerConn == nil {
		outputChan <- "Error: You are not connected to the server."
		return

	}
	// Şifreleme İşlemleri
	currentPass := cmdHandler.RoomPasswords[cmdHandler.CurrentRoom] // Mevcut şifre
	encryptedText := security.EncryptMessage(rawInput, currentPass) // kilitle

	msg := encryptedText + "\n"

	// Ağa Yazma İşlemi
	_, writeErr := cmdHandler.ServerConn.Write([]byte(msg))
	if writeErr != nil {
		outputChan <- "Message could not be sent !!"
		// Mesaj gönderilemedi!!
	} else {
		myMsg := fmt.Sprintf("ME| %s", rawInput)
		outputChan <- myMsg
	}
}
