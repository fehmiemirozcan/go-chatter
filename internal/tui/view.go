package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- YENİ BEYİN: İÇERİKLERİ MOTORLARA YÜKLEYEN FONKSİYON ---
// Bu fonksiyon orijinal motorların (viewport) içini metinle doldurur.
func (m model) syncViewports() model {
	if m.width == 0 {
		return m
	}

	chatWidth := (m.width * 70) / 100

	// mesaj karekter sınırılamsı 30 karaterden sonra alt satıra insin sağa veya sola yaslı şekilde

	// -- Chat İçeriği --
	var chatContentBuilder strings.Builder

	for _, msg := range m.message {
		var formattedMsg string

		if msg.IsSys {
			if strings.Contains(msg.Content, "Welcome") {
				formattedMsg = welcomeStyle.Width(chatWidth).Align(lipgloss.Center).Render(msg.Content)
			} else if strings.Contains(msg.Content, "/help") {
				helpBox := helpBoxStyle.Render(msg.Content)
				formattedMsg = lipgloss.NewStyle().Width(chatWidth - 4).Align(lipgloss.Left).Render(helpBox)
			} else if strings.Contains(msg.Content, "whisper") {
				whisperColor := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Italic(true).Render(msg.Content)
				formattedMsg = lipgloss.NewStyle().Width(chatWidth - 4).Align(lipgloss.Right).Render(whisperColor)
			} else {
				formattedMsg = systemMessageStyle.Width(chatWidth - 4).Render(msg.Content)
			}
		} else if msg.Isme {

			// Kullanıcı ve zaman ksıımlarını ayarlıyorum
			prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render(fmt.Sprintf("[%s] %s:", msg.Timestamp, m.userName))

			// ayarladığım ksıısmlarla mesajı birlştiriyorum tek metin haline getiriyorum
			fullText := fmt.Sprintf("%s %s", prefix, msg.Content)

			contentWidth := chatWidth - 4
			chatHalf := contentWidth / 2

			realWidth := lipgloss.Width(fullText)

			// Mesaj ekranın yarısını geçerse, tam ortada kilitlenir ve alt satıra iner!
			if realWidth > chatHalf {
				realWidth = chatHalf
			}

			// Mesaj sol duvara sıfır yanaşmasın, 10 karakter nefes payı kalsın
			maxBubbleWidth := chatWidth - 10
			if realWidth > maxBubbleWidth {
				realWidth = maxBubbleWidth
			}

			// Metini sola yaslıyorum
			innerBlock := lipgloss.NewStyle().Width(realWidth).Align(lipgloss.Left).Render(fullText)

			// Ekranın sol boş kısmını oluşturuyorum
			formattedMsg = lipgloss.NewStyle().
				Width(contentWidth).
				Align(lipgloss.Right).
				Render(innerBlock)

		} else if msg.IsWhisper {

			prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render(fmt.Sprintf("[%s] Whisper from %s:", msg.Timestamp, msg.Sender))
			fulText := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Italic(true).Render(fmt.Sprintf("%s %s", prefix, msg.Content))

			contentWidth := chatWidth - 4
			chatHalf := contentWidth / 2

			realWidth := lipgloss.Width(fulText)
			if realWidth > chatHalf {
				realWidth = chatHalf
			}

			formattedMsg = lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left).Render(fulText)

		} else {

			userColor := getUserColor(msg.Sender)

			prefix := lipgloss.NewStyle().Foreground(lipgloss.Color(userColor)).Bold(true).Render(fmt.Sprintf("[%s] %s:", msg.Timestamp, msg.Sender))
			fulText := fmt.Sprintf("%s %s", prefix, msg.Content)

			contentWidth := chatWidth - 4
			chatHalf := contentWidth / 2

			realWidth := lipgloss.Width(fulText)
			if realWidth > chatHalf {
				realWidth = chatHalf
			}
			innerBlock := lipgloss.NewStyle().Width(realWidth).Align(lipgloss.Left).Render(fulText)

			formattedMsg = lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left).Render(innerBlock)
		}
		chatContentBuilder.WriteString(formattedMsg + "\n")
	}
	m.vp.SetContent(chatContentBuilder.String())
	return m
}

// -- ANA EKRANI ÇİZEN FONKSİYON --
func (m model) View() string {
	if m.width < 95 || m.height < 20 {
		return "Please enlarge the terminal screen."
	}
	if m.width == 0 {
		return "Loading..."
	}
	// Ölçüler
	chatWidth := (m.width * 70) / 100
	rightWidth := m.width - chatWidth

	mainHeight := m.height - 7

	// chatbox boyunda 2 tane sağ panael olacak ve araalrında 1'er boşluk olacak
	rightInternalHeight := mainHeight - 2
	roomsHeight := rightInternalHeight / 2
	usersHeight := rightInternalHeight - roomsHeight

	// Odak ve Dinamik Renkler
	activeColor := lipgloss.Color("42")   // Parlak yeşil
	passiveColor := lipgloss.Color("238") // Koyu gri

	chatBorderColor := passiveColor
	roomBorderColor := passiveColor
	userBorderColor := passiveColor
	inputBorderColor := passiveColor

	// Hangi kutudaysak onu aktif yapıyoruz
	switch m.focusBox {
	case FocusInput:
		inputBorderColor = activeColor
	case FocusChat:
		chatBorderColor = activeColor
	case FocusRooms:
		roomBorderColor = activeColor
	case FocusUsers:
		userBorderColor = activeColor
	}

	// Ekran Parçaları Çizme
	chatBox := chatBoxStyle.BorderForeground(chatBorderColor).Width(chatWidth - 2).Height(mainHeight).Render(m.vp.View())

	roomsPanel := roomsBoxStyle.BorderForeground(roomBorderColor).Width(rightWidth - 2).Height(roomsHeight).Render(m.roomList.View())
	usersPanel := usersBoxStyle.BorderForeground(userBorderColor).Width(rightWidth - 2).Height(usersHeight).Render(m.userList.View())

	inputBar := inputBoxStyle.BorderForeground(inputBorderColor).Width(m.width - 2).Render(m.ti.View())

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).MarginLeft(2).
		Render(" [ / ] Search   [ Tab ] Change   [ Esc ] Back   [ Ctrl+C ] Quit ")

	// Parçaları Birleştirme
	rightColumn := lipgloss.JoinVertical(lipgloss.Top, roomsPanel, usersPanel)
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, chatBox, rightColumn)

	return lipgloss.JoinVertical(lipgloss.Left, mainArea, inputBar, footer, "")
}

// ODALAR LİSTESİ
type roomDelegate struct{}

func (d roomDelegate) Height() int                               { return 1 }
func (d roomDelegate) Spacing() int                              { return 0 }
func (d roomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d roomDelegate) Render(w io.Writer, l list.Model, index int, listItem list.Item) {
	i, ok := listItem.(roomItem)
	if !ok {
		return
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true) // Beyaz
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))               // Yeşil
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))              // Gri

	var renderedString string
	if index == l.Index() {
		renderedString = selectedStyle.Render("-> " + i.name)
	} else if i.isActive {
		renderedString = activeStyle.Render(". " + i.name)
	} else {
		renderedString = normalStyle.Render(" " + i.name)
	}
	fmt.Fprint(w, renderedString)
}

// KULLANICI LİSTESİ
type userDelegate struct{}

func (d userDelegate) Height() int                               { return 1 }
func (d userDelegate) Spacing() int                              { return 0 }
func (d userDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d userDelegate) Render(w io.Writer, l list.Model, index int, listItem list.Item) {
	i, ok := listItem.(userItem)
	if !ok {
		return
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	meStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("80")) // mevcut kullanıcı rengi
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var statusText string
	switch i.status {
	case "online":
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("online") // Yeşil
	case "away":
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("away") // Sarı
	case "dnd":
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("dnd") // Kırmızı
	default:
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("offline") // Çevrimdışı/Bilinmiyor
	}

	displayStr := fmt.Sprintf("%s %s", i.name, statusText)

	var renderedString string
	if index == l.Index() {
		renderedString = selectedStyle.Render("-> " + displayStr)
	} else if i.isMe {
		renderedString = meStyle.Render(". " + displayStr)
	} else {
		renderedString = normalStyle.Render(" " + displayStr)
	}
	fmt.Fprint(w, renderedString)
}
