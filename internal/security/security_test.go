package security

import (
	"encoding/base64"
	"testing"
)

func TestEncryptMessage_Password(t *testing.T) {
	msg := "Hello World"
	password := ""

	result := EncryptMessage(msg, password)

	if result != msg {
		t.Errorf("Expected:%q , Received:%q", password, result)
	}
}
func TestEncryptAndDecryptMessages(t *testing.T) {
	msg := "Hello World!"
	password := "123"

	encryptedText := EncryptMessage(msg, password)

	tooShortBase64 := base64.StdEncoding.EncodeToString([]byte("short"))

	tests := []struct {
		testName      string
		encryptedText string
		password      string
		expected      string
	}{
		{"Boş password", msg, "", msg},
		{"Geçersiz base64", "base64_degil", "123", "base64_degil"},
		{"Verinin kisa olmasi", tooShortBase64, "123", tooShortBase64},
		{"Yanlis sifre", encryptedText, "000", "The messages could not be deciphered."},
		{"Basarili durum", encryptedText, password, msg},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			decyptText := DecryptMessage(test.encryptedText, test.password)
			if decyptText != test.expected {
				t.Errorf("[%s] Expected:%q , Received:%q", test.testName, test.expected, decyptText)
			}
		})
	}

}
