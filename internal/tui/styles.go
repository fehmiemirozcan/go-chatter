package tui

import "github.com/charmbracelet/lipgloss"

var (
	// İlk açılıştaki "Welcome to the secret chat"
	welcomeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Align(lipgloss.Center)
	// Sunucu mesajları - parlak sarı
	systemMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Align(lipgloss.Left)

	// /help kutusu
	helpBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("250")).Padding(1, 2).Align(lipgloss.Left)
	// Ekranın %75'ini kaplayan ana mesajlaşma - mor/mavi
	chatBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	// Ekranın en altı Input Bar - Koyu gri
	inputBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Align(lipgloss.Left).PaddingLeft(1)
	// Sağ üst panelin çerçevesi (Odalar Listesini sarmalar)
	roomsBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	// Sağ alt panelin çerçevesi (Kullanıcı Listesini sarmalar)
	usersBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
)

var (
	colorPalette = []lipgloss.Color{
		lipgloss.Color("39"),  // Mavi
		lipgloss.Color("20"),  // Turuncu
		lipgloss.Color("118"), // Açık Yeşil
		lipgloss.Color("161"), // Yakut Rengi
		lipgloss.Color("129"), // Mor
		lipgloss.Color("6"),   // Turkuaz
	}
)

func getUserColor(userName string) lipgloss.Color {
	sum := 0
	for _, char := range userName {
		sum += int(char)
	}
	return colorPalette[sum%len(colorPalette)] //1
}
