package tui

import (
	"sort" // Odaları sabitlemek için eklendi
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case RoomUpdateMsg:
		m.rooms = msg.Rooms
		m.activeRoom = msg.ActiveRoom
		m = m.syncViewports()

		// Odaları alfabetik sırala
		sort.Slice(msg.Rooms, func(i, j int) bool {
			return strings.ToLower(msg.Rooms[i]) < strings.ToLower(msg.Rooms[j])
		})

		var items []list.Item
		for _, roomData := range msg.Rooms {
			if roomData != "" {
				parts := strings.Split(roomData, ":")
				roomName := parts[0]
				locked := len(parts) == 2 && parts[1] == "1"
				items = append(items, roomItem{
					name:     roomName,
					isActive: roomName == m.activeRoom,
					isLocked: locked,
				})
			}
		}

		// Eğer kullanıcı o an arama yapmıyorsa listeyi güncelle (Aramasını bölme)
		cmd = m.roomList.SetItems(items)
		cmds = append(cmds, cmd, ListenMessage(m.outChan, m.userName))

	case RoomUsers:
		m.users = msg.Users
		m = m.syncViewports()

		// Kullanıcıları alfabetik sırala
		sort.Slice(msg.Users, func(i, j int) bool {
			return strings.ToLower(msg.Users[i]) < strings.ToLower(msg.Users[j])
		})

		var items []list.Item
		for _, user := range msg.Users {
			if user != "" {
				parts := strings.Split(user, ":")
				name := parts[0]
				status := "online"
				if len(parts) > 1 {
					status = parts[1]
				}
				items = append(items, userItem{name: name, isMe: name == m.userName, status: status})
			}
		}

		cmd = m.userList.SetItems(items)
		cmds = append(cmds, ListenMessage(m.outChan, m.userName))

	case CustomMsg:
		if msg.OutsideChat.IsSys && strings.HasPrefix(msg.OutsideChat.Content, "[SYSTEM] Successfully moved to room") {
			m.message = nil // Eski odanın mesajlarını sil
		}
		m.message = append(m.message, msg.OutsideChat)
		m = m.syncViewports()
		m.vp.GotoBottom()
		cmds = append(cmds, ListenMessage(m.outChan, m.userName))

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		chatWidth := (m.width * 70) / 100
		rightWidth := m.width - chatWidth
		mainHeight := m.height - 7

		rightInternalHeight := mainHeight - 2
		roomsHeight := rightInternalHeight / 2
		usersHeight := rightInternalHeight - roomsHeight

		m.vp.Width = chatWidth - 2
		m.vp.Height = mainHeight

		m.viewRooms.Width = rightWidth - 4
		m.viewRooms.Height = roomsHeight
		m.roomList.SetSize(m.viewRooms.Width, m.viewRooms.Height)

		m.viewUsers.Width = rightWidth - 4
		m.viewUsers.Height = usersHeight
		m.userList.SetSize(m.viewUsers.Width, m.viewUsers.Height)

		m = m.syncViewports()
		m.vp.GotoBottom()

	case tea.MouseMsg:
		chatWidth := (m.width * 70) / 100
		mainHeight := m.height - 7
		roomsHeight := (mainHeight - 2) / 2

		inInputPane := msg.Y >= mainHeight && msg.Y < mainHeight+5
		inHelpBar := msg.Y == mainHeight+5

		inRoomsPane := msg.X >= chatWidth && msg.Y <= roomsHeight+2 && msg.Y < mainHeight
		inUsersPane := msg.X >= chatWidth && msg.Y > roomsHeight+2 && msg.Y < mainHeight
		inChatPane := msg.X < chatWidth && msg.Y < mainHeight

		// 1. Tıklama ile Odak Değişimi
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if inHelpBar {
				if msg.X >= 2 && msg.X <= 14 {
					// [ / ] Search
					return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
				} else if msg.X >= 18 && msg.X <= 30 {
					// [ Tab ] Focus
					return m.Update(tea.KeyMsg{Type: tea.KeyTab})
				} else if msg.X >= 35 && msg.X <= 47 {
					// [ Esc ] Back
					return m.Update(tea.KeyMsg{Type: tea.KeyEsc})
				} else if msg.X >= 50 && msg.X <= 64 {
					// [ Ctrl+C ] Quit
					return m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
				}
			} else if inRoomsPane {
				m.focusBox = FocusRooms
				m.ti.Blur()
				clickedIndex := msg.Y - 3
				if clickedIndex >= 0 && clickedIndex < len(m.roomList.VisibleItems()) {
					m.roomList.Select(clickedIndex)
					return m.Update(tea.KeyMsg{Type: tea.KeyEnter})
				}
				return m, nil
			} else if inUsersPane {
				m.focusBox = FocusUsers
				m.ti.Blur()
				clickedIndex := msg.Y - roomsHeight - 5
				if clickedIndex >= 0 && clickedIndex < len(m.userList.VisibleItems()) {
					m.userList.Select(clickedIndex)
					return m.Update(tea.KeyMsg{Type: tea.KeyEnter})
				}

				return m, nil
			} else if inChatPane {
				m.focusBox = FocusChat
				m.ti.Blur()
			} else if inInputPane {
				m.focusBox = FocusInput
				m.ti.Focus()
			}
		}

		// 2. Scroll (Kaydırma) İşlemleri
		isWheelUp := msg.Button == tea.MouseButtonWheelUp
		isWheelDown := msg.Button == tea.MouseButtonWheelDown

		if isWheelUp {
			if inRoomsPane {
				m.roomList.CursorUp()
			} else if inUsersPane {
				m.userList.CursorUp()
			} else if inChatPane {
				m.vp.LineUp(3)
			} else if inInputPane {
			}
			return m, nil
		}
		if isWheelDown {
			if inRoomsPane {
				m.roomList.CursorDown()
			} else if inUsersPane {
				m.userList.CursorDown()
			} else if inChatPane {
				m.vp.LineDown(3)
			}
			return m, nil
		}

	case tea.KeyMsg:
		// Programdan Çıkış
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Başka bir tuşa basılırsa oto-tamamlama hafızasını sıfırla
		if msg.String() != "tab" {
			m.candidatesList = []string{}
			m.candidateIndex = 0
		}

		if msg.String() == "tab" {
			val := m.ti.Value()
			// Oto-tamamlama (Arama çubuğu aktif ve / ile başlıyorsa)
			if m.focusBox == FocusInput && strings.HasPrefix(val, "/") {
				allCommands := []string{"/help", "/createroom", "/join", "/whisper", "/status", "/quit"}
				if len(m.candidatesList) == 0 {
					m.candidateIndex = 0
					m.originalCandidate = val
					for _, command := range allCommands {
						if strings.HasPrefix(command, m.originalCandidate) {
							m.candidatesList = append(m.candidatesList, command)
						}
					}
				} else {
					if len(m.candidatesList) > 0 {
						m.candidateIndex = (m.candidateIndex + 1) % len(m.candidatesList)
					}
				}
				if len(m.candidatesList) > 0 {
					m.ti.SetValue(m.candidatesList[m.candidateIndex])
					m.ti.SetCursor(len(m.ti.Value()))
				}
				return m, nil
			}

			// Oto-tamamlama yoksa paneller arası geçiş yap
			m.focusBox = (m.focusBox + 1) % 4
			if m.focusBox == FocusInput {
				m.ti.Focus()
			} else {
				m.ti.Blur()
			}
			return m, nil
		}

		if msg.String() == "esc" {
			if m.focusBox == FocusRooms && m.roomList.FilterState() != list.Unfiltered {
				m.roomList.ResetFilter()
				return m, nil
			}
			if m.focusBox == FocusUsers && m.userList.FilterState() != list.Unfiltered {
				m.userList.ResetFilter()
				return m, nil
			}
			m.focusBox = FocusInput
			m.ti.Focus()
			return m, nil
		}

		switch m.focusBox {
		case FocusRooms:
			if msg.String() == "enter" {
				targetRoom := ""
				isLocked := false

				// İmleç kaymasını tamamen bypass edip, yazdığı kelimeyi birebir buluyoruz!
				if m.roomList.FilterState() != list.Unfiltered {
					filterText := strings.TrimSpace(m.roomList.FilterValue())
					for _, it := range m.roomList.Items() {
						if r, ok := it.(roomItem); ok && strings.EqualFold(r.name, filterText) {
							targetRoom = r.name
							isLocked = r.isLocked
							break
						}
					}
				}

				// Eğer yazdığı şey tam eşleşmediyse, imlecin o an üzerinde olduğu odayı al
				if targetRoom == "" {
					if sel, ok := m.roomList.SelectedItem().(roomItem); ok {
						targetRoom = strings.TrimSpace(sel.name)
						isLocked = sel.isLocked
					}
				}

				if targetRoom != "" {
					m.roomList.ResetFilter()
					if isLocked {
						m.ti.SetValue("/join " + targetRoom + " ")
						m.ti.Focus()
						m.focusBox = FocusInput
					} else {
						go func(r string) { m.inChan <- "/join " + r }(targetRoom)
						m.ti.Reset()
						m.ti.Focus()
						m.focusBox = FocusInput
					}
				}
			} else {
				// Enter değilse, harfleri listeye gönder (Filtreleme çalışsın)
				m.roomList, cmd = m.roomList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case FocusUsers:
			if msg.String() == "enter" {
				targetUser := ""
				if m.userList.FilterState() != list.Unfiltered {
					filterText := strings.TrimSpace(m.userList.FilterValue())
					for _, it := range m.userList.Items() {
						if u, ok := it.(userItem); ok && strings.EqualFold(u.name, filterText) {
							targetUser = u.name
							break
						}
					}
				}
				if targetUser == "" {
					if sel, ok := m.userList.SelectedItem().(userItem); ok {
						targetUser = strings.TrimSpace(sel.name)
					}
				}

				if targetUser != "" {
					m.userList.ResetFilter()
					m.ti.SetValue("/whisper " + targetUser + " :" + " ")
					m.ti.Focus()
					m.focusBox = FocusInput
				}
			} else {
				m.userList, cmd = m.userList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case FocusInput:
			m.ti, cmd = m.ti.Update(msg)
			if msg.String() == "enter" {
				val := m.ti.Value()
				if strings.TrimSpace(val) != "" {
					m.history = append(m.history, val)
					m.historyIndex = len(m.history)
					m.inChan <- val
					m.ti.SetValue("")
					m.vp.GotoBottom()
				}
			} else if msg.String() == "up" {
				// YUKARI OK: Geçmiş mesajlara git
				if m.historyIndex > 0 {
					m.historyIndex--
					m.ti.SetValue(m.history[m.historyIndex])
					m.ti.SetCursor(len(m.ti.Value())) // İmleci sona al
				}
			} else if msg.String() == "down" {
				// AŞAĞI OK: Daha yeni mesajlara git
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					m.ti.SetValue(m.history[m.historyIndex])
				} else {
					// En sona gelindiyse inputu boşalt
					m.historyIndex = len(m.history)
					m.ti.SetValue("")
				}
				m.ti.SetCursor(len(m.ti.Value()))
			}
			cmds = append(cmds, cmd)
		}

	default:
		// Kütüphanenin arka plan işlemlerini (Filtreleme, İmleç vb.) yutmamak için
		// gelen tüm bilinmeyen mesajları listelere iletmeliyiz!
		m.roomList, cmd = m.roomList.Update(msg)
		cmds = append(cmds, cmd)

		m.userList, cmd = m.userList.Update(msg)
		cmds = append(cmds, cmd)

		m.vp, cmd = m.vp.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func ListenMessage(outChan chan string, userName string) tea.Cmd {
	return func() tea.Msg {
		rawMsg := <-outChan
		if rawMsg == "SYS_EXIT_NOW" {
			return tea.QuitMsg{}
		}

		var msg ChatMessage
		msg.Timestamp = time.Now().Format("15:04:05")

		if strings.HasPrefix(rawMsg, "ROOMS|") {
			content := strings.TrimPrefix(rawMsg, "ROOMS|")
			parts := strings.SplitN(content, "|", 2)
			var rooms []string
			var active string
			if len(parts) == 2 {
				rooms = strings.Split(parts[0], ",")
				active = parts[1]
			}
			return RoomUpdateMsg{Rooms: rooms, ActiveRoom: active}

		} else if strings.HasPrefix(rawMsg, "ME|") {
			msg.Isme = true
			msg.Content = strings.TrimPrefix(rawMsg, "ME|")

		} else if strings.HasPrefix(rawMsg, "MSG|") {
			msg.Isme = false
			msg.IsSys = false
			content := strings.TrimPrefix(rawMsg, "MSG|")
			parts := strings.SplitN(content, ":", 2)
			if len(parts) == 2 {
				msg.Sender, msg.Content = parts[0], strings.TrimSpace(parts[1])
			} else {
				msg.Content = content
			}

		} else if strings.HasPrefix(rawMsg, "USERS|") {
			content := strings.TrimPrefix(rawMsg, "USERS|")
			var users []string
			if content != "" {
				users = strings.Split(content, ",")
			}
			return RoomUsers{Users: users}

		} else if strings.HasPrefix(rawMsg, "WHISPER|") {
			msg.IsWhisper = true
			parts := strings.Split(rawMsg, "|")
			if len(parts) == 3 {
				msg.Sender, msg.Content = parts[1], strings.TrimSpace(parts[2])
			}
		} else {
			msg.Isme, msg.IsSys, msg.Content = false, true, rawMsg
		}

		msg.Content = strings.TrimSpace(msg.Content)
		return CustomMsg{OutsideChat: msg}
	}
}
