package client

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Racemir/go-chatter/internal/command"
)

func TestStartChat_SplitBrainRace(t *testing.T) {
	var (
		originalDiscoverLeaderFunc = DiscoverLeaderFunc
		originalStartUDPFunc       = StartUDPFunc
		originalDialFunc           = DialFunc
		originalStartServerFunc    = StartServerFunc
	)

	defer func() {
		DiscoverLeaderFunc = originalDiscoverLeaderFunc
		StartUDPFunc = originalStartUDPFunc
		DialFunc = originalDialFunc
		StartServerFunc = originalStartServerFunc
	}()

	// Kimse bulunamadı simülasyonu
	DiscoverLeaderFunc = func(port string) (string, error) {
		return "", fmt.Errorf("No one was found.")
	}
	StartUDPFunc = func(port string) {}

	// İlk gelen lider, ikinci gelen hata alır!
	var portLock sync.Mutex
	var portInUse bool

	// Yeni ekledim kimin lider olduğunu buradan görüyorum
	StartServerFunc = func() error {
		portLock.Lock()
		defer portLock.Unlock()

		if portInUse {
			return fmt.Errorf("bind: address already in use") // Adres zaten kullanımda
		}
		portInUse = true
		return nil
	}

	DialFunc = func(network string, address string) (net.Conn, error) {
		mockClient, mockServer := net.Pipe()
		go func() {
			buf := make([]byte, 1024)
			mockServer.Read(buf)
			mockServer.Write([]byte("Successful")) // Doğru yazım!
		}()
		return mockClient, nil
	}

	var wg sync.WaitGroup
	wg.Add(2)

	users := []string{"Emir", "Barış"}
	results := make(chan string, 2)

	for _, user := range users {
		go func(userName string) {
			defer wg.Done()
			inputChan := make(chan string, 1)
			outputChan := make(chan string, 10)
			deathSignal := make(chan struct{})

			clientConn, serverConn := net.Pipe()
			serverConn.Close() // Fişi çekip yarışı başlatıyoruz

			go StartChat(userName, clientConn, inputChan, outputChan, deathSignal)

			<-outputChan // Welcome to the secret chat
			<-outputChan // Help mesajı

			var reaction string //
			for i := 0; i < 2; i++ {
				take := <-outputChan
				if strings.Contains(take, "You are the leader") || strings.Contains(take, "Another client claimed") || strings.Contains(take, "Connection rejected by new leader.") {
					reaction = take
				}
			}
			results <- fmt.Sprintf("%s: %s", userName, strings.TrimSpace(reaction))
			clientConn.Close()
		}(user)
	}

	wg.Wait()
	close(results)

	leaderCount := 0 //lider sayısı
	for res := range results {
		if strings.Contains(res, "You are the leader") {
			leaderCount++
		}
	}

	if leaderCount > 1 {
		// split brain problemi oldu sohbet odasında birden fazla lider gözüküyor
		t.Errorf("Aynı anda lider olmaya çalışan birden fazla kişi var! Lider Sayısı: %d", leaderCount)
	} else if leaderCount == 0 {
		t.Errorf("Kimse lider olamadı!") // sistem kilitlenmiş yada tamamen çoktü
	} else {
		t.Log("Harika! Sadece bir kişi lider oldu.")
	}
}

func TestStartChat_HostMigration(t *testing.T) {
	var (
		originalDiscoverLeaderFunc = DiscoverLeaderFunc
		originalStartUDPFunc       = StartUDPFunc
		originalDialFunc           = DialFunc
		originalStartServerFunc    = StartServerFunc
	)

	defer func() {
		DiscoverLeaderFunc = originalDiscoverLeaderFunc
		StartUDPFunc = originalStartUDPFunc
		DialFunc = originalDialFunc
		StartServerFunc = originalStartServerFunc
	}()

	tests := []struct {
		testName     string
		expectedChan string
	}{
		{"New Leader Found", "[!] New leader found at 192.168.1.100. Connecting..."},
		{"I Am the Leader", "[!] You are the leader! Starting server..."},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			inputChan := make(chan string, 1)
			outputChan := make(chan string, 10)
			deathSignal := make(chan struct{})

			clientConn, serverConn := net.Pipe()
			serverConn.Close() // Fişi çekiyoruz

			StartUDPFunc = func(port string) {}
			StartServerFunc = func() error { return nil }

			if test.testName == "I Am the Leader" {
				DiscoverLeaderFunc = func(port string) (string, error) {
					return "", fmt.Errorf("No one was found.")
				}
			} else {
				DiscoverLeaderFunc = func(port string) (string, error) {
					return "192.168.1.100", nil
				}
			}

			// Her dial çağrısında yeni bir pipe oluşturuyorum
			DialFunc = func(network string, address string) (net.Conn, error) {
				mockClient, mockServer := net.Pipe()
				go func() {
					buf := make([]byte, 1024)
					mockServer.Read(buf)
					// DİKKAT: "Successful" kelimesindeki yazım hatası düzeltildi!
					mockServer.Write([]byte("Successful"))
				}()
				return mockClient, nil
			}

			go StartChat("TestUser", clientConn, inputChan, outputChan, deathSignal)

			<-outputChan // "Welcome to the secret chat"
			<-outputChan // Help mesajı
			<-outputChan // Leader connection lost

			// Zaman ayarlı okuma (Test sonsuza kadar takılmasın diye)
			select {
			case take := <-outputChan:
				if !strings.Contains(take, test.expectedChan) {
					t.Errorf("Expected to contain: %q , Received: %q", test.expectedChan, take)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("Test zaman aşımına uğradı! Beklenen mesaj gelmedi: %s", test.expectedChan)
			}

			clientConn.Close()
		})
	}
}

func TestStartChat_MigrationFailures(t *testing.T) {
	var (
		originalDiscoverLeaderFunc = DiscoverLeaderFunc
		originalStartUDPFunc       = StartUDPFunc
		originalDialFunc           = DialFunc
		originalStartServerFunc    = StartServerFunc
	)

	defer func() {
		DiscoverLeaderFunc = originalDiscoverLeaderFunc
		StartUDPFunc = originalStartUDPFunc
		DialFunc = originalDialFunc
		StartServerFunc = originalStartServerFunc
	}()

	tests := []struct {
		testName      string
		isLeaderFound bool   // ağda lider var mı
		serverStarts  bool   // sunucu başladı mı
		serverReply   string // serverdan gelen cevap
		expectedChan  string // hata mesajı
	}{
		{"ağda lider var ama bağlanamiyor", true, false, "UnSuccesful", "[!] Migration failed: Could not connect to remote leader."},
		{"lider ben yada baskasi degil", false, false, "UnSuccesful", "[!] Migration failed: Connection rejected by new leader."},
		{"lider kendi serverina oda kuramiyor", false, true, "UnSuccesful", "[!] You are the leader! Starting server..."},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			inputChan := make(chan string, 1)
			outputChan := make(chan string, 10)
			deatSignal := make(chan struct{})

			clientConn, serverConn := net.Pipe()
			serverConn.Close()

			StartUDPFunc = func(port string) {}
			if test.isLeaderFound {
				DiscoverLeaderFunc = func(port string) (string, error) {
					return "192.168.1.100", nil
				}
			} else {
				DiscoverLeaderFunc = func(port string) (string, error) {
					return "", fmt.Errorf("No one was found.")
				}
			}

			if test.serverStarts {
				StartServerFunc = func() error { return nil }
			} else {
				StartServerFunc = func() error { return fmt.Errorf("bind: adress already in use") }
			}

			DialFunc = func(network string, address string) (net.Conn, error) {
				mockClient, mocServer := net.Pipe()
				go func() {
					buf := make([]byte, 1024)
					mocServer.Read(buf)
					mocServer.Write([]byte(test.serverReply)) // UnSuccesful
				}()
				return mockClient, nil
			}

			go StartChat("TestUser", clientConn, inputChan, outputChan, deatSignal)

			<-outputChan // "Welcome to the secret chat"
			<-outputChan // Help mesajı
			<-outputChan // Leader connection lost

			found := false
			timeout := time.After(2 * time.Second)

		loop:
			for {
				select {
				case take := <-outputChan:
					if strings.Contains(take, test.expectedChan) {
						found = true
						break loop
					}
				case <-timeout:
					break loop
				}
			}
			if !found {
				t.Errorf("[%s] Expected final failure message: %q", test.testName, test.expectedChan)
			}
			clientConn.Close()
		})
	}
}
func TestStartChat_InputAndDeathSignal(t *testing.T) {

	tests := []struct {
		testName     string
		rawInput     string // klavyeden gönderilecek msg
		triggerDeath bool
		expectedChan string
	}{
		{"Null Data", " ", false, ""},
		{"Incorrect Command", "/wrong_commant", false, "Unknown command. Type /help"},
		{"Standard message", "selam", false, "ME| selam"},
		{"DeathSignal", "", true, "Connection completely lost"},
	}

	for _, test := range tests {
		inputChan := make(chan string, 1)
		outputChan := make(chan string, 10)
		deathSignal := make(chan struct{})
		clientConn, serverConn := net.Pipe()

		go func() {
			buf := make([]byte, 1024)
			for {
				_, readErr := serverConn.Read(buf)
				if readErr != nil {
					break
				}
			}
		}()

		go StartChat("TestUser", clientConn, inputChan, outputChan, deathSignal)

		<-outputChan // "Welcome to the secret chat"
		<-outputChan // Help mesajı

		if test.triggerDeath == true {
			deathSignal <- struct{}{}
		} else {
			inputChan <- test.rawInput
		}

		if test.testName == "Null Data" {
			continue
		}

		take := <-outputChan
		if !strings.Contains(take, test.expectedChan) && take != test.expectedChan {
			t.Errorf("[%s] Expected to contain: %q, Received: %q", test.testName, test.expectedChan, take)
		}

		clientConn.Close()
		serverConn.Close()
	}
}
func TestParseIncomingMessage(t *testing.T) {
	// msg boş
	// msg başında ROOMCHANGE|SUCCESS|
	// msg başında ROOMCHANGE|ERR|
	// msg başında ALLROOMS|
	// msg başında ROOMUSERS|
	// msg başında WHISPER|
	// parts 2 ye eşitse yani msg 2 parçaya ayrılmışsa
	// mesaj içeriğinde : yoksa sistem mesajı olacak

	tests := []struct {
		testName string
		sent     string
		expected string //kanaldan beklenen mesaj

	}{
		{"Boş Veri", "", ""},
		{"Oda Değişikliği", "ROOMCHANGE|SUCCESS|", "[SYSTEM] Successfully moved to room ''!\n"},
		{"Hatalı Oda Değişikliği", "ROOMCHANGE|ERR|", "[ERROR] Room action failed: \n"},
		{"Oda Listesi", "ALLROOMS|", "ROOMS||"},
		{"Kullanıcı Listesi", "ROOMUSERS|", "USERS|"},
		{"Whisper Mesaj", "WHISPER|emir|merahaba", "WHISPER|emir|merahaba"},
		{"Msg 2 Ve ':' var", "emir: merhaba", "MSG| emir: merhaba\n"},
		{"Msg İçinde ':' Yok", "emir merhaba", "emir merhaba"},
	}

	for _, test := range tests {
		chanParsa := make(chan string, 1)
		cmdHand := &command.Handler{UserName: "emir"}

		parseIncomingMessage(test.sent, cmdHand, chanParsa)
		if test.expected != "" {
			take := <-chanParsa
			if take != test.expected {
				t.Errorf("Network Error ->> Expected:%q , Received:%q", test.expected, take)
			}
		}
	}
}

// func TestCleanInput(t *testing.T) {
// 	// mesaj gönderip deneriz bu kısmı
// 	tests := []struct {
// 		sent         string
// 		Expected     []string
// 		inputControl bool
// 	}{
// 		{" /join  denemeler   ", []string{"/join", "denemeler"}, true},
// 		{" /join DeNeMeLeR", []string{"/join", "denemeler"}, true},
// 		{"/join  DeNeMeLeR   ", []string{"/join", "denemeler"}, true},
// 		{"   /jOIN     DeNeMeLeR  ", []string{"/join", "denemeler"}, true},
// 		{" /JOIN  denemeler", []string{"/join", "denemeler"}, true},
// 		{" /join  DENEMELER", []string{"/join", "denemeler"}, true},
// 	}

//		for _, test := range tests {
//			cleanOutput := cleanInput(test.sent)
//			if !slices.Equal(cleanOutput, test.Expected) {
//				t.Errorf("Expected:%v , Received:%v", test.Expected, cleanOutput)
//			}
//		}
//	}
func TestHandleUserCommand(t *testing.T) {

	// liste 0 ise
	// commandName /quit ise bağlanbtı kapat
	// commanddictionary false ise kanala "Unknown command. Type /help"
	tests := []struct {
		testName      string
		sent          string
		expected      string
		cloeseNetwork bool
	}{
		{"Boş Veri", "", "", false},
		{"/quit Geldiyse", "/quit", "", true}, // dikkat ilk deatlock
		{"/join Geldiyse", "/join", "Please specify a room name. Example: /join <room_name> <password>", false},
		{"Yanlış Komut Geldiyse", "/komut", "Unknown command. Type /help", false},
	}
	for _, test := range tests {
		cahnUserCom := make(chan string, 2)
		clientConn, serverConn := net.Pipe()

		hand := &command.Handler{ServerConn: serverConn, UserName: "TestUser"}

		handleUserCommand(test.sent, hand, cahnUserCom)

		if test.cloeseNetwork == true {
			_, writeErr := clientConn.Write([]byte("hello"))
			if writeErr == nil {
				t.Error("The /quit command was executed, but the network connection was not closed!")
			}
		}

		if test.expected != "" {
			take := <-cahnUserCom
			if test.expected != take {
				t.Errorf("Network Error ->> Expected:%q , Received:%q", test.expected, take)
			}
		}
	}
}

func TestSendNormalMessage(t *testing.T) {
	// saniye sınırı
	// bağlantı kopması
	// write bir sorunu
	// başarılı senaryo
	tests := []struct {
		testName           string
		sent               string
		expectedNetworkMsg string
		closeConn          bool
		spam               bool
	}{
		{"Spam Kontrolü", "merhaba", "SYSTEM| You're sending messages too quickly!", false, true},
		{"Bağlantı Kontrolü", "merhaba", "Error: You are not connected to the server.", false, false},
		{"Ağ Yazma Hatası", "merhaba", "Message could not be sent !!", true, false},
		{"Mesaj Gönderildi", "merhaba", "ME| merhaba", false, false},
	}
	for _, test := range tests {
		chanNormalMsg := make(chan string, 1)

		lastMsgTime := time.Now().Add(-1 * time.Hour)
		if test.spam == true {
			lastMsgTime = time.Now()
		}

		clientConn, serverConn := net.Pipe()

		if test.closeConn == true {
			serverConn.Close()
			clientConn.Close()
		}

		if test.testName == "Bağlantı Kontrolü" {
			serverConn = nil
		}

		if test.testName == "Mesaj Gönderildi" {
			go func() {
				buf := make([]byte, 1024)
				clientConn.Read(buf)

			}()
		}

		hand := &command.Handler{ServerConn: serverConn}
		sendNormalMessage(test.sent, hand, chanNormalMsg, &lastMsgTime)

		take := <-chanNormalMsg
		if take != test.expectedNetworkMsg {
			t.Errorf("Expected:%q , Received:%q", test.expectedNetworkMsg, take)
		}
	}
}
