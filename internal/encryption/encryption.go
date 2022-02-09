package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Encryption struct {
	Key []byte
}

func (e *Encryption) Encrypt(text string) (string, error) {
	logSnippet := "[internal][encryption][Encrypt] =>"
	fmt.Printf("%s (textToEncrypt): '%s'", logSnippet, text)

	plainText := []byte(text)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	fmt.Printf("%s (encryptedText): '%s'", logSnippet, base64.URLEncoding.EncodeToString(cipherText))

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func (e *Encryption) Decrypt(encryptedText string) (string, error) {
	logSnippet := "[internal][encryption][Decrypt] =>"
	fmt.Printf("%s (encryptedText): '%s'", logSnippet, encryptedText)

	cipherText, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("invalid length - cannot decrpt")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	fmt.Printf("%s (decryptedText): '%s'", logSnippet, fmt.Sprintf("%s", cipherText))

	return fmt.Sprintf("%s", cipherText), nil
}
