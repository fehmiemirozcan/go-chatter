package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// usurname kısmı için fonksiyon
// en fazla 1 kelimden oluşan bir kullanıcı adı
// Daha sonrası için <emir> Join gibi bir şey de olabilir

func ClientWelcome() (string, error) {

	fmt.Print("Enter a one-word username\n")
	for {
		fmt.Print("> ")
		userName := GetInput()
		if len(userName) == 0 {
			fmt.Print("You should have entered a username.\n")
			continue
		}
		if strings.Contains(userName, " ") {
			fmt.Print("I'm sorry, your username should have been a single word.\n")
			continue
		}

		fmt.Printf("Welcome %s\n", userName)
		return userName, nil
	}
}

// msjlaşmak için bir fonksiyon
func GetInput() string {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	line := scanner.Text()
	line = strings.TrimSpace(line)
	return line
}
