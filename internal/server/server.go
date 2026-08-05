package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Messenger interface {
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	Close() error
}

type Client struct {
	Conn   Messenger
	Name   string
	Status string
}

type Room struct {
	Users    map[Messenger]*Client
	Password string
	Owner    string
	Mutex    sync.Mutex
}

type RouterHub struct {
	Rooms         map[string]*Room // Oda ismi ve adresi.
	IPConnections map[string]int   // Aynı Ip da kaç kişi bağlı
	Mutex         sync.Mutex
}

// Sunucu açıldığında boş bir merkez olacak.
func NewRouterHub() *RouterHub {
	return &RouterHub{
		Rooms:         make(map[string]*Room),
		IPConnections: make(map[string]int),
	}
}

func (h *RouterHub) StartServer(port string) error {
	list, ListErr := net.Listen("tcp", port)
	if ListErr != nil {
		//fmt.Println("Master Server could not start:")
		// Ana sunucu başlatılamadı.
		return ListErr
	}

	//fmt.Println("--> Master Server listening on port", port)
	// Ana sunucu belirtilen portta dinleme yapıyor.
	go func() {
		defer list.Close()
		for {
			conn, acceptErr := list.Accept()
			if acceptErr != nil {
				continue
			}
			go h.handleConnection(conn)
		}
	}()
	return nil
}

// Kimlik ve oda kontrolleri.
func (h *RouterHub) handleConnection(conn net.Conn) {

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
	h.Mutex.Lock()
	if h.IPConnections[ip] >= 3 {
		h.Mutex.Unlock()
		conn.Write([]byte("SYSTEM|Connection limit reached for this IP!\n"))
		conn.Close()
		return
	}
	h.IPConnections[ip]++
	h.Mutex.Unlock()

	buf := make([]byte, 1024)
	c, readErr := conn.Read(buf)
	if readErr != nil {
		conn.Close()
		return
	}

	// format = "OdaAdı|Şifre|İsim" şeklinde olacak.
	incomingData := string(buf[:c])
	parts := strings.Split(incomingData, "|")

	if len(parts) < 4 {
		conn.Write([]byte("Invalide Handshake Format"))
		// Geçersiz El Sıkışma Formatı.
		conn.Close()
		return
	}

	action := parts[0] // createroom , join
	if action != "createroom" && action != "join" {
		conn.Write([]byte("Invalid action"))
		conn.Close()
		return
	}
	roomName := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])
	userName := strings.TrimSpace(parts[3])

	// Böyle bir oda var mı ? yoksa oluşturulacak.
	h.Mutex.Lock()

	// Global İsim Çakışma Kontrolü
	nameTaken := false
	for _, r := range h.Rooms {
		r.Mutex.Lock()
		for _, existingClient := range r.Users {
			// EqualFold ile Eda == eda == EDA eşleşmesini yakalıyoruz
			if isEqualName(existingClient.Name, userName) {
				nameTaken = true
			}
		}
		r.Mutex.Unlock()
		if nameTaken {
			break
		}
	}

	// Eğer isim sunucunun herhangi bir odasında kullanılıyorsa bağlantıyı kes
	if nameTaken {
		h.Mutex.Unlock()
		conn.Write([]byte("Username is already taken by another user!"))
		conn.Close()
		return
	}

	// Oda Yönetimi
	room, exists := h.Rooms[roomName]

	if action == "join" && !exists {
		h.Mutex.Unlock()
		conn.Write([]byte("Room not found!"))
		// Oda bulunamadı.
		h.Mutex.Lock()
		h.IPConnections[ip]--
		h.Mutex.Unlock()
		conn.Close()
		return
	}

	if action == "createroom" && exists && roomName != "global" {
		h.Mutex.Unlock()
		conn.Write([]byte("Room already exists!"))
		// Oda zaten mevcut!
		conn.Close()
		return
	}

	if !exists {
		ownerName := userName
		if roomName == "global" {
			ownerName = "SYSTEM" // Global odayı sistemin üzerine yapıyoruz
		}
		room = &Room{
			Users:    make(map[Messenger]*Client),
			Password: password,
			Owner:    ownerName,
		}
		h.Rooms[roomName] = room

	}

	h.Mutex.Unlock()
	room.Mutex.Lock()

	// Şifre kontrolü.
	if room.Password != "" && room.Password != password {
		room.Mutex.Unlock()
		conn.Write([]byte("Wrong Password"))
		// Yanlış şifre
		conn.Close()
		return
	}
	/*
		// isim kullanılıyor mu ?
		for _, existingName := range room.Users {
			if existingName == userName {
				room.Mutex.Unlock()
				conn.Write([]byte("Unsuccessful"))
				conn.Close()
				return
			}
		}
	*/
	room.Users[conn] = &Client{Name: userName, Status: "online"}
	room.Mutex.Unlock()

	conn.Write([]byte("Successful"))

	// Tcp paketi birleşiyor biraz bekleteceğim
	time.Sleep(100 * time.Millisecond)

	// Odadaki herkese güncel kullanıcı listesini yolla
	h.broadcastRoomUsers(room)

	// Ağa yeni bir oda eklendiyse (veya biri girdiyse) güncel oda listesini herkese yolla
	h.broadcastAllRoomsList()

	joinMsg := fmt.Sprintf("%s [+] joined the room.\n", userName)
	h.broadcastToRoom(room, conn, joinMsg)

	// chati dinlemeye devam et
	h.listenClientMessages(conn, room, userName)
}

// Bir kullanıcının bağlantısının kopması, yazılan verinin yakalanması.
func (h *RouterHub) listenClientMessages(conn net.Conn, room *Room, userName string) {
	defer func() {
		room.Mutex.Lock()
		delete(room.Users, conn)
		room.Mutex.Unlock()

		h.broadcastRoomUsers(room) // Yeni liste
		time.Sleep(100 * time.Millisecond)

		leaveMsg := fmt.Sprintf("%s [!] left the room.\n", userName)
		h.broadcastToRoom(room, conn, leaveMsg)

		clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
		h.Mutex.Lock()
		h.IPConnections[clientIP]-- // Çıkan kişinin hakkını geri ver
		h.Mutex.Unlock()
		conn.Close()
	}()

	lastMsgTime := time.Now().Add(-1 * time.Second)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		rawMsg := scanner.Text()
		cleanMsg := strings.TrimSpace(rawMsg)
		if cleanMsg == "" {
			continue
		}
		now := time.Since(lastMsgTime)
		if now < 300*time.Millisecond {
			conn.Write([]byte("SYSTEM| You're sending messages too quickly!\n"))
			continue
		}
		lastMsgTime = time.Now()

		// Durum değiştirme
		if strings.HasPrefix(cleanMsg, "STATUS|") {
			parts := strings.Split(cleanMsg, "|")
			if len(parts) >= 2 {
				newStatus := strings.TrimSpace(parts[1])
				room.Mutex.Lock()
				if client, ok := room.Users[conn]; ok {
					client.Status = newStatus
				}
				room.Mutex.Unlock()
				h.broadcastRoomUsers(room)
			}
			continue
		}

		// Oda değiştirme ve oda kurma
		if strings.HasPrefix(cleanMsg, "CMD|") {
			parts := strings.Split(cleanMsg, "|")
			if len(parts) >= 3 {
				action := parts[1] // join veya createroom
				targetRoomName := strings.TrimSpace(parts[2])
				targetPassword := ""
				if len(parts) == 4 {
					targetPassword = strings.TrimSpace(parts[3])
				}

				h.Mutex.Lock()
				targetRoom, exists := h.Rooms[targetRoomName]

				// Katılmak istenen oda yoksa
				if action == "join" && !exists {
					h.Mutex.Unlock()
					conn.Write([]byte("ROOMCHANGE|ERR|Room not found!\n"))
					continue
				}

				// Kurulmak istenen oda zaten varsa
				if action == "createroom" && exists && targetRoomName != "global" {
					h.Mutex.Unlock()
					conn.Write([]byte("ROOMCHANGE|ERR|Room already exists!\n"))
					continue
				}

				if action == "createroom" && !exists {
					userRoomCount := 0
					for _, r := range h.Rooms {
						if r.Owner == userName { // Bu kullanıcıya ait kaç oda var
							userRoomCount++
						}
					}

					// Bir kullanıcı en fazla 3 oda kurabilir (ilk giriş yapan sadece 2 oda kurabilir globa oda onun sayılır.)
					if userRoomCount >= 3 {
						h.Mutex.Unlock()
						conn.Write([]byte("ROOMCHANGE|ERR|You can only create up to 3 rooms!\n"))
						continue
					}
				}

				// Eski odadan kullanıcının bağlantısını siliyoruz
				room.Mutex.Lock()
				movingClient := room.Users[conn]
				delete(room.Users, conn)
				room.Mutex.Unlock()

				// Eski odadakilere kullanıcının ayrıldığını duyuruyoruz
				h.broadcastRoomUsers(room)
				leaveMsg := fmt.Sprintf("%s [!] left the room.\n", userName)
				h.broadcastToRoom(room, conn, leaveMsg)

				// Eğer oda createroom ile yeni oluşturuluyorsa bellekte yer açıyoruz
				if !exists {

					targetRoom = &Room{
						Users:    make(map[Messenger]*Client),
						Password: targetPassword,
						Owner:    userName,
					}
					h.Rooms[targetRoomName] = targetRoom
				}
				h.Mutex.Unlock()

				// Yeni odanın şifresini kontrol ediyoruz
				targetRoom.Mutex.Lock()
				if targetRoom.Password != "" && targetRoom.Password != targetPassword {
					targetRoom.Mutex.Unlock()

					// Şifre yanlışsa eski odasına geri yerleştiriyoruz.
					room.Mutex.Lock()
					room.Users[conn] = movingClient
					room.Mutex.Unlock()

					h.broadcastRoomUsers(room)
					conn.Write([]byte("ROOMCHANGE|ERR|Wrong Password!\n"))
					continue
				}

				// Kullanıcıyı yeni odaya başarıyla kaydediyoruz
				targetRoom.Users[conn] = movingClient
				targetRoom.Mutex.Unlock()

				// Fonksiyon parametresi olan room değişkenini güncelliyoruz.
				// Böylece hem döngüdeki normal mesajlar yeni odaya gider hem de
				// kullanıcı çıkış yaparsa defer fonksiyonu doğru odayı temizler.
				room = targetRoom

				// İstemciye odanın başarıyla değiştiğini bildir
				// iyoruz
				conn.Write([]byte(fmt.Sprintf("ROOMCHANGE|SUCCESS|%s\n", targetRoomName)))

				time.Sleep(100 * time.Millisecond)

				// Yeni odadaki herkesin ekranını ve global oda listelerini tazeliyoruz
				h.broadcastRoomUsers(room)
				h.broadcastAllRoomsList()

				joinMsg := fmt.Sprintf("%s [+] joined the room.\n", userName)
				h.broadcastToRoom(room, conn, joinMsg)
			}
			continue
		}

		// Whisper mesaj atma işlemi
		if strings.HasPrefix(cleanMsg, "WHISPER|") {
			senderName := userName
			args := strings.SplitN(cleanMsg, "|", 3)

			if len(args) == 3 {
				targetName := args[1]
				message := args[2]

				h.whisperedMessage(senderName, targetName, message)
			}
			continue
		}

		//fmt.Println("\n[SUNUCU LOGU] Encrypted Package Being Transported:", msg)
		formattedMsg := fmt.Sprintf("%s: %s", userName, cleanMsg)

		h.broadcastToRoom(room, conn, formattedMsg)

	}
}

// Mesaj gönderen dışındaki kullanıcılara msj gönderir.
func (h *RouterHub) broadcastToRoom(room *Room, sender Messenger, message string) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	for clientConn := range room.Users {
		if clientConn != sender {
			clientConn.Write([]byte(message))
		}
	}
}

// Yeni oda oluşturulduğunda tüm kullanıcıların Rooms panelini güncel tutur.
func (h *RouterHub) broadcastGlobalToRoom(message string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	for _, room := range h.Rooms {
		room.Mutex.Lock()
		for conn := range room.Users {
			conn.Write([]byte(message))
		}
		room.Mutex.Unlock()
	}
}

func (h *RouterHub) broadcastRoomUsers(room *Room) {
	room.Mutex.Lock()

	// Odadaki kullanıcılar
	var userDatas []string
	for _, client := range room.Users {
		userDatas = append(userDatas, fmt.Sprintf("%s:%s", client.Name, client.Status))
	}

	// Virgülle ayır bir tag koy
	userListStr := strings.Join(userDatas, ",")
	broadcastMsg := fmt.Sprintf("ROOMUSERS|%s\n", userListStr)

	for conn := range room.Users {
		conn.Write([]byte(broadcastMsg))
	}

	room.Mutex.Unlock()
}

func (h *RouterHub) broadcastAllRoomsList() {
	h.Mutex.Lock()

	// Odaların isimlerini topla
	var roomNames []string
	for name, room := range h.Rooms {
		isLocked := "0"
		if room.Password != "" {
			isLocked = "1"
		}
		roomData := fmt.Sprintf("%s:%s", name, isLocked)
		roomNames = append(roomNames, roomData)
	}

	h.Mutex.Unlock()

	// Virgülle ayır bir tag koy
	roomListStr := strings.Join(roomNames, ",")
	broadcastMsg := fmt.Sprintf("ALLROOMS|%s\n", roomListStr)

	// tüm odalara yolla
	h.broadcastGlobalToRoom(broadcastMsg)

}

func (h *RouterHub) whisperedMessage(senderName, targetName, message string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	// böyle bir user var mı kontrolü

	for _, room := range h.Rooms {
		room.Mutex.Lock()

		for userConn, client := range room.Users {

			if isEqualName(client.Name, targetName) {

				whisperMsg := fmt.Sprintf("WHISPER|%s|%s\n", senderName, message)
				userConn.Write([]byte(whisperMsg))

				room.Mutex.Unlock()
				return
			}
		}
		room.Mutex.Unlock()
	}
}

func normalizeForComparison(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "İ", "i")
	s = strings.ReplaceAll(s, "I", "ı")
	s = strings.ToLower(s)
	replacer := strings.NewReplacer(
		"ç", "c",
		"ğ", "g",
		"ı", "i",
		"ö", "o",
		"ş", "s",
		"ü", "u",
		"i̇", "i",
	)
	return replacer.Replace(s)
}

func isEqualName(name1, name2 string) bool {
	return normalizeForComparison(name1) == normalizeForComparison(name2)
}
