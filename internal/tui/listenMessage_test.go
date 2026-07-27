package tui

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestListenMessage(t *testing.T) {
	// Test edeceğimiz kullanıcının adı
	mockUserName := "TestUser"

	tests := []struct {
		name     string
		rawInput string  // Kanaldan gelecek ham metin
		expected tea.Msg // Fonksiyonun dönmesini beklediğimiz Bubbletea mesaj nesnesi
	}{
		{
			name:     "Acil Çıkış Sinyali",
			rawInput: "SYS_EXIT_NOW",
			expected: tea.QuitMsg{},
		},
		{
			name:     "Eksik Ayırıcılı ROOMS (Uç Senaryo)",
			rawInput: "ROOMS|global",
			// len(parts) == 2 koşulu sağlanmadığı için diziler boş döner
			expected: RoomUpdateMsg{Rooms: nil, ActiveRoom: ""},
		},
		{
			name:     "İki Noktası Olmayan MSG (Fallback Senaryosu)",
			rawInput: "MSG|selam naber",
			expected: CustomMsg{
				OutsideChat: ChatMessage{
					Isme:      false,
					IsSys:     false,
					Sender:    "",            // İki nokta olmadığı için gönderen bulunamadı
					Content:   "selam naber", // Metnin tamamı içeriğe düştü
					Timestamp: "",            // Test hilesi için boş bırakıyoruz
				},
			},
		},
	}

	for _, test := range tests {
		// Kapasitesi 1 olan kanal yaratıyoruz (Goroutine kullanmaya gerek kalmıyor)
		outChan := make(chan string, 1)

		// Ham veriyi kanala atıyoruz
		outChan <- test.rawInput

		// Fonksiyonu çağırıp bize dönen tea.Cmd nesnesini (fonksiyonu) alıyoruz
		cmdFunc := ListenMessage(outChan, mockUserName)

		// cmdFunc'ı çalıştırarak içindeki tea.Msg nesnesini ürettiriyoruz
		gotMsg := cmdFunc()

		// Eğer dönen mesaj CustomMsg ise zaman damgasını uçuruyoruz
		if customGot, ok := gotMsg.(CustomMsg); ok {
			customGot.OutsideChat.Timestamp = ""
			gotMsg = customGot // Değiştirilmiş halini geri atıyoruz
		}

		// Beklenen nesne ile üretilen nesneyi derinlemesine karşılaştırıyoruz
		if !reflect.DeepEqual(gotMsg, test.expected) {
			t.Errorf("[%s] Expected: %+v , Received: %+v", test.name, test.expected, gotMsg)
		}
	}
}
