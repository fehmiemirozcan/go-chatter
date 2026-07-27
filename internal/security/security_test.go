package security

import "testing"

func TestEncryptAndDecryptMessages(t *testing.T) {
	msg := "Hello World!"
	password := "123"
	wrongPassword := "456"

	encryptedText := EncryptMessage(msg, password)
	decyptText := DecryptMessage(encryptedText, password)

	if msg != decyptText {
		t.Error("The original text could not be encrypted.")
	}
	encryptedMsg := EncryptMessage(msg, password)
	decyptMsg := DecryptMessage(encryptedMsg, wrongPassword)
	if msg == decyptMsg {
		t.Error("The text was translated using the wrong password.")
	}
}
