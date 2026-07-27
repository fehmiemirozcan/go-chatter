package server

import (
	"net"
	"strings"
	"testing"
	"time"
)

type MockClient struct {
	MessagesReceived []string
}

func (m *MockClient) Write(b []byte) (n int, err error) {
	m.MessagesReceived = append(m.MessagesReceived, string(b))
	return len(b), nil
}

func (m *MockClient) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockClient) Close() error {
	return nil
}

func TestBroadcastToRoom(t *testing.T) {
	sender := &MockClient{}
	client1 := &MockClient{}
	client2 := &MockClient{}

	TestRoom := &Room{
		Users: make(map[Messenger]*Client),
	}

	TestRoom.Users[sender] = &Client{Name: "sender"}
	TestRoom.Users[client1] = &Client{Name: "client_Emir"}
	TestRoom.Users[client2] = &Client{Name: "client_Gece"}

	hub := NewRouterHub()
	senderMessage := "Love Forever"

	hub.broadcastToRoom(TestRoom, sender, senderMessage)

	if len(sender.MessagesReceived) != 0 {
		t.Error("A message was received by the sender.")
	}

	if len(client1.MessagesReceived) != 1 {
		t.Error("The message did not reach client1.")
	} else if client1.MessagesReceived[0] != senderMessage {
		t.Errorf("The wrong message was sent. Expected:%s , Received:%s", senderMessage, client1.MessagesReceived[0])
	}

	if len(client2.MessagesReceived) != 1 {
		t.Error("The message did not reach client2.")
	} else if client2.MessagesReceived[0] != senderMessage {
		t.Errorf("The wrong message was sent. Expected:%s , Received:%s", senderMessage, client2.MessagesReceived[0])
	}

}

func TestBroadcastGlobalToRoom(t *testing.T) {

	// 2 tane test odası oluşturucaz
	TestRoom0 := &Room{
		Users: make(map[Messenger]*Client),
	}

	TestRoom1 := &Room{
		Users: make(map[Messenger]*Client),
	}

	// 3 tane kullanıcı oluşturucaz

	Emir := &MockClient{}
	Nurullah := &MockClient{}
	Eda := &MockClient{}

	// 2 kullanıcı bir odaya 1 kullanıcı diğer odaya yerleştiricez

	TestRoom0.Users[Emir] = &Client{Name: "Emir"}
	TestRoom0.Users[Nurullah] = &Client{Name: "Nurullah"}
	TestRoom1.Users[Eda] = &Client{Name: "eda"}

	// broadcastGlobalToRoom fonksiyonumu çalıştırıcam
	hub := NewRouterHub()
	hub.Rooms["oda1"] = TestRoom0
	hub.Rooms["oda2"] = TestRoom1

	sentMessage := "ALLROOMS|global,oda1,oda2"
	hub.broadcastGlobalToRoom(sentMessage)
	// messageReceived == 1 olmalı ve msj doğru olmalı

	if len(Emir.MessagesReceived) != 1 {
		t.Error("The room list is not updating.")
	} else if Emir.MessagesReceived[0] != sentMessage {
		t.Errorf("Expected:%s, Received: %s", sentMessage, Emir.MessagesReceived[0])
	}

	if len(Nurullah.MessagesReceived) != 1 {
		t.Error("The room list is not updating.")
	} else if Nurullah.MessagesReceived[0] != sentMessage {
		t.Errorf("Expected: %s, Received:%s", sentMessage, Nurullah.MessagesReceived[0])
	}

	if len(Eda.MessagesReceived) != 1 {
		t.Error("The room list is not updating.")
	} else if Eda.MessagesReceived[0] != sentMessage {
		t.Errorf("Expected:%s, Received: %s", sentMessage, Eda.MessagesReceived[0])
	}
}

func TestWhisperedMessage(t *testing.T) {
	// 3 tane kullanıcı oluştur
	user1 := &MockClient{}
	targetUser := &MockClient{}
	senderUser := &MockClient{}

	// 1 tane oda oluşturucaz
	TestRoom := &Room{
		Users: make(map[Messenger]*Client),
	}

	// kullanıcıları odalara ekle
	TestRoom.Users[user1] = &Client{Name: "user1"}
	TestRoom.Users[targetUser] = &Client{Name: "targetUser"}
	TestRoom.Users[senderUser] = &Client{Name: "senderUser"}

	// hub ve fonksiyonumuz çalıştırıcaz
	hub := NewRouterHub()
	hub.Rooms["whisperRoom"] = TestRoom

	whisperMessage := "merhaba"
	expectedMessage := "WHISPER|senderUser|merhaba\n"
	hub.whisperedMessage("senderUser", "targetUser", whisperMessage)
	// sadece istenen kullanıcıya mesajın gitmesi gerekli

	if len(senderUser.MessagesReceived) != 0 {
		t.Error("The person who sent the whisper message received their own message.")
	}

	if len(targetUser.MessagesReceived) != 1 {
		t.Error("Whisper message not sent.")
	} else if targetUser.MessagesReceived[0] != expectedMessage {
		t.Errorf("Expected:%q , Received:%q", expectedMessage, targetUser.MessagesReceived[0])
	}

	if len(user1.MessagesReceived) != 0 {
		t.Error("You sent a whisper message to the wrong user.")
	}
}

func TestBroadcastRoomUser(t *testing.T) {
	// 2 tane kullanıcı oluştur
	user0 := &MockClient{}
	user1 := &MockClient{}

	// oda oluştur
	TestRoom := &Room{
		Users: make(map[Messenger]*Client),
	}

	// kullancııları odaya yerleştir
	TestRoom.Users[user0] = &Client{Name: "user0", Status: "online"}
	TestRoom.Users[user1] = &Client{Name: "user1", Status: "away"}

	// hub ve fonskiyon çalıştır
	hub := NewRouterHub()

	hub.broadcastRoomUsers(TestRoom)
	// map için de emir:online var mı ? araması

	if len(user0.MessagesReceived) != 1 {
		t.Error("Status update message could not be sent.")
	} else if !strings.Contains(user0.MessagesReceived[0], "user0:online") {
		t.Error("incorrect status definition")
	}

	if len(user1.MessagesReceived) != 1 {
		t.Error("Status update message could not be sent.")
	} else if !strings.Contains(user1.MessagesReceived[0], "user1:away") {
		t.Error("incorrect status definition")
	}
}

func TestIsEqualName(t *testing.T) {

	tests := []struct {
		userName0 string
		userName1 string
		expected  bool
	}{
		{"EMIR", "emir", true},
		{"Gece", "emir", false},
		{"IŞIK", "ısık", true},
		{"Yağ", "yag", true},
	}

	for _, test := range tests {
		boolNormalize := isEqualName(test.userName0, test.userName1)

		if boolNormalize != test.expected {
			t.Errorf("Expected:%v , Received:%v", test.expected, boolNormalize)
		}
	}
}

func TestNewRouterHub(t *testing.T) {

	hub := NewRouterHub()
	if hub == nil {
		t.Error("Hub object could not be created.")
	}

	if hub.Rooms == nil {
		t.Error("Rooms could not be created.")
	}
}

func TestHandleConnection_InvalidHandshake(t *testing.T) {
	hub := NewRouterHub()
	clientConn, serverConn := net.Pipe()
	go hub.handleConnection(serverConn)
	clientConn.Write([]byte("createroom"))

	buf := make([]byte, 1024)
	s, readErr := clientConn.Read(buf)
	if readErr != nil {
		t.Error("Read rejected.")
	}
	incMsg := string(buf[:s])

	expectedError := "Invalide Handshake Format"
	if incMsg != expectedError {
		t.Errorf("Expected: %q, Received: %q", expectedError, incMsg)
	}

}

func TestHandleConnection_Errors(t *testing.T) {

	hub := NewRouterHub()

	TestRoom := &Room{
		Users:    make(map[Messenger]*Client),
		Password: "123",
	}
	hub.Rooms["TestRoom"] = TestRoom
	user0 := &MockClient{}
	TestRoom.Users[user0] = &Client{Name: "emir"}

	tests := []struct {
		testName       string
		goingMsg       string
		expectedAnswer string
	}{
		{"Gecersiz Aksiyon", "hackle|TestRoom|123|Gece", "Invalid action"},
		{"Olmayan Odaya Katılma", "join|Test1|12|Gece", "Room not found!"},
		{"Yanlış Şifre", "join|TestRoom|12|Gece", "Wrong Password"},
		{"İsim Çakışması", "join|TestRoom|123|Emir", "Username is already taken by another user!"},
		{"Mevcut Odayı Kurma", "createroom|TestRoom|123|Gece", "Room already exists!"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			clientConn, serverConn := net.Pipe()
			go hub.handleConnection(serverConn)

			clientConn.Write([]byte(test.goingMsg))

			buf := make([]byte, 1024)
			c, clientEror := clientConn.Read(buf)
			if clientEror != nil {
				t.Error("Read rejected.")
			}
			msgs := string(buf[:c])
			if msgs != test.expectedAnswer {
				t.Errorf("TestName: %q, Expected:%s, Received:%s", test.testName, test.expectedAnswer, msgs)
			}
		})
	}
}

func TestListenClientMessages_StatusCommand(t *testing.T) {
	hub := NewRouterHub()

	TestRoom := &Room{
		Users: make(map[Messenger]*Client),
	}
	hub.Rooms["TestRoom"] = TestRoom

	clientConn, serverConn := net.Pipe()

	testClient := &Client{
		Name:   "Emir",
		Status: "online",
	}
	TestRoom.Users[serverConn] = testClient

	go hub.listenClientMessages(serverConn, TestRoom, testClient.Name)

	time.Sleep(time.Millisecond * 350)

	_, writeErr := clientConn.Write([]byte("STATUS|dnd\n"))
	if writeErr != nil {
		t.Errorf("Writing Error: %v", writeErr)
	}

	time.Sleep(time.Millisecond * 50)

	if testClient.Status != "dnd" {
		t.Errorf("Status did not change! Expected:dnd, Received:%q", testClient.Status)
	}

}

func TestListenClientMessages_CmdCommand(t *testing.T) {

	tests := []struct {
		testName       string
		goingCommand   string
		expectedAnswer string
	}{

		{"Başarılı Oda Kurma", "CMD|createroom|oda1|123\n", "ROOMCHANGE|SUCCESS|oda1\n"},

		{"Var Olan Odayı Kurmaya Çalışma", "CMD|createroom|deneme|123\n", "ROOMCHANGE|ERR|Room already exists!\n"},
		{"Olmayan Odaya Katılma", "CMD|join|hayaletOda|123\n", "ROOMCHANGE|ERR|Room not found!\n"},
		{"Yanlış Şifre ile Katılma", "CMD|join|deneme|yanlisSifre\n", "ROOMUSERS|Emir:\n"},

		// kullanıcı bağlantısı silindi mi ?
		// kullanıcı ayrıldı mesjı
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			hub := NewRouterHub()

			TestRoom := &Room{
				Users:    make(map[Messenger]*Client),
				Password: "123",
			}
			hub.Rooms["deneme"] = TestRoom

			clientTest := &Client{Name: "Emir"}

			clientConn, serverConn := net.Pipe()

			TestRoom.Users[serverConn] = clientTest

			go hub.listenClientMessages(serverConn, TestRoom, clientTest.Name)

			time.Sleep(time.Millisecond * 350)

			//serverden gelen mesjı okuma işlem iyapıp kontrol edeceğim
			clientConn.Write([]byte(test.goingCommand))

			time.Sleep(time.Millisecond * 50)

			buf := make([]byte, 1024)
			n, _ := clientConn.Read(buf)
			receivedAnswer := string(buf[:n])
			if receivedAnswer != test.expectedAnswer {
				t.Errorf("Expected: %q, Received: %q", test.expectedAnswer, receivedAnswer)
			}
		})
	}
}

func TestListenClientMessages_WhisperCommand(t *testing.T) {

	// 3 tane user oluşturucaz
	hub := NewRouterHub()

	TestRoom := &Room{
		Users: make(map[Messenger]*Client),
	}
	hub.Rooms["TestRoom"] = TestRoom

	user0 := &Client{Name: "user0"}
	user1 := &Client{Name: "user1"}
	user2 := &Client{Name: "user2"}

	clientConn0, serverConn0 := net.Pipe()
	clientConn1, serverConn1 := net.Pipe()
	clientConn2, serverConn2 := net.Pipe()

	TestRoom.Users[serverConn0] = user0
	TestRoom.Users[serverConn1] = user1
	TestRoom.Users[serverConn2] = user2

	go hub.listenClientMessages(serverConn0, TestRoom, user0.Name)

	time.Sleep(time.Millisecond * 350)
	// whisper mesaj atacak user0 -> user1 'e
	clientConn0.Write([]byte("WHISPER|user1|merhaba\n"))

	buf := make([]byte, 1024)
	s, _ := clientConn1.Read(buf)
	msg := string(buf[:s])
	if msg != "WHISPER|user0|merhaba\n" {
		t.Error("The message content is incroccet.")
	}

	// user2 ' ye mesaj gitmeyecek
	buff := make([]byte, 1024)
	clientConn2.SetReadDeadline(time.Now().Add(time.Millisecond * 50))
	_, readErr := clientConn2.Read(buff)
	if readErr == nil {
		t.Error("The message went to the wrong person.")
	}
}

func TestStartServer(t *testing.T) {
	hub := NewRouterHub()
	go hub.StartServer(":8080")
	time.Sleep(time.Millisecond * 50)
	conn, dialErr := net.Dial("tcp", "localhost:8080")
	if dialErr != nil {
		t.Error("")
	}
	conn.Close()
}
