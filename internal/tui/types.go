package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

const (
	FocusInput = 0
	FocusChat  = 1
	FocusRooms = 2
	FocusUsers = 3
)

type model struct {
	width        int
	height       int
	historyIndex int // dizinin neresindeyiz

	ti textinput.Model

	history []string // gönderilen msjları tutacağız
	message []ChatMessage

	vp        viewport.Model
	viewRooms viewport.Model
	viewUsers viewport.Model

	userName string
	inChan   chan string
	outChan  chan string

	rooms      []string
	users      []string
	activeRoom string

	focusBox int

	candidatesList    []string
	candidateIndex    int
	originalCandidate string

	roomList list.Model
	userList list.Model
}

type RoomUpdateMsg struct {
	Rooms      []string
	ActiveRoom string
}

type RoomUsers struct {
	Users []string
}

type ChatMessage struct {
	Sender    string // mesajı kim gönderdi ?
	Content   string // mesajın veya komutun içeriği ?
	Timestamp string // mesajın geldiği saat
	Isme      bool   // mesajı gönderen ben miyim ?
	IsSys     bool   // sistem mesajı mı ?
	IsWhisper bool   // fısıltı mesjı mı ?

}

type CustomMsg struct {
	OutsideChat ChatMessage
}

type roomItem struct {
	name     string
	isActive bool
	isLocked bool
}

func (i roomItem) Title() string       { return i.name }
func (i roomItem) Description() string { return "" }
func (i roomItem) FilterValue() string { return i.name }

type userItem struct {
	name   string
	isMe   bool
	status string
}

func (i userItem) Title() string       { return i.name }
func (i userItem) Description() string { return "" }
func (i userItem) FilterValue() string { return i.name }
