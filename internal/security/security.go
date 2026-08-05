package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

// Kullanıcının girdiği şifreyi AES 32 bytelık anahtara çevirme
func createHash(key string) []byte {
	privatePassword := sha256.New()
	privatePassword.Write([]byte(key))
	return privatePassword.Sum(nil)
}

// Msg kitle
func EncryptMessage(plainText string, password string) string {

	if password == "" {
		return plainText
		// Oda şifresizse msj kitlemiyoruz.
	}

	key := createHash(password)

	// AES-256 bloğu oluştur. Gerçek şifreyi üretiyor
	block, _ := aes.NewCipher(key)

	// GCM aktif ediyoruz. Veriyi etiketler değişmesi durumunda anlamak
	// 128 bit boyuta sahip
	gcm, _ := cipher.NewGCM(block)

	// Benzersiz bir nonce üret
	nonce := make([]byte, gcm.NonceSize()) // 12 byte boş yer ayır
	io.ReadFull(rand.Reader, nonce)        // 12 bytı rastgele sayılarla doldurur (aynı msjı 2 defa yolladığımızda şifreler aynı olmasın diye)

	cipherTextBytes := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	// [12 Byte Rastgele Nonce] + [Şifrelenmiş Mesaj Byte'ları] + [16 Byte Doğrulama Etiketi]
	// Seal(hedefDizi, kullanilacakNonce, mesaj, ekstraVeri)

	return base64.StdEncoding.EncodeToString(cipherTextBytes)

}

// Msg kilidi çöz
func DecryptMessage(encryptedText, password string) string {

	// Oda şifresizse çözecek msj yok
	if password == "" {
		return encryptedText
	}

	cipherTextBytes, decodeStringErr := base64.StdEncoding.DecodeString(encryptedText)
	if decodeStringErr != nil {
		return encryptedText
		// Msg 64 bit değilse aynen yaz gönder
	}

	key := createHash(password)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonceSize := gcm.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return encryptedText
		// Veri kısaysa şifreli değildir aynen yaz
	}

	nonce, cipherText := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]
	// İlk 12 byte ve 12 byte dan sonrası (şifreli msg ve MAC etiketi)

	plainTextBytes, openErr := gcm.Open(nil, nonce, cipherText, nil) // şifreli metni düz metne çevir
	// İlk nil çıkan sonucu yeni bir alana koy. İkinci nil ekstra veri yok
	if openErr != nil {
		return "The messages could not be deciphered."
		// Mesajlar çözülemedi
	}

	return string(plainTextBytes)
}
