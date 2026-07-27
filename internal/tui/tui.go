package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// uygulamanın çalıştırılması için başlangıç ​​komutunu döndüren bir fonksiyondur.
func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, ListenMessage(m.outChan, m.userName))
}

func initialModel(in, out chan string, userName string) model {

	vp := viewport.New(0, 0)
	roomVp := viewport.New(0, 0)
	usersVp := viewport.New(0, 0)

	ti := textinput.New()
	ti.Prompt = " > "
	ti.Placeholder = "Secret chat..."
	ti.ShowSuggestions = true
	ti.CharLimit = 280
	ti.SetSuggestions([]string{"/help", "/createroom", "/join", "/whisper", "/status", "/quit"})
	ti.Focus()

	m := model{
		width:        0,
		height:       0,
		historyIndex: 0,

		ti: ti,

		history: []string{},
		message: []ChatMessage{},

		vp:        vp,
		viewRooms: roomVp,
		viewUsers: usersVp,

		userName: userName,
		inChan:   in,
		outChan:  out,

		rooms:      []string{},
		users:      []string{},
		activeRoom: "",

		focusBox: FocusInput,

		candidatesList:    []string{},
		candidateIndex:    0,
		originalCandidate: "",
	}

	// ODALAR
	m.roomList = list.New([]list.Item{}, roomDelegate{}, 30, 24)
	m.roomList.Title = "ROOMS"
	m.roomList.SetShowHelp(false)
	m.roomList.SetShowStatusBar(false)
	m.roomList.SetShowPagination(false)

	// KULLANICILAR
	m.userList = list.New([]list.Item{}, userDelegate{}, 30, 24)
	m.userList.Title = "USERS"
	m.userList.SetShowHelp(false)
	m.userList.SetShowStatusBar(false)
	m.userList.SetShowPagination(false)

	return m
}

func StartUI(inChan, outChan chan string, userName string) error {
	p := tea.NewProgram(initialModel(inChan, outChan, userName), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, runErr := p.Run()
	return runErr
}
