package zoom

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

// ZoomContextDecrypter ...
// context - Encrypted Zoom App Contex (x-zoom-zapp-context)
// secretKey - Client Secret Key.
func ContextDecrypter(context string, secretKey []byte) []byte {
	ciphertext, err := base64.RawURLEncoding.DecodeString(context)
	decoder := bytes.NewReader(ciphertext)
	ivLength := make([]byte, 1)
	_, _ = decoder.Read(ivLength)
	iv := make([]byte, ivLength[0])
	_, _ = decoder.Read(iv)
	aadLengthBytes := make([]byte, 2)
	_, _ = decoder.Read(aadLengthBytes)
	addLength := binary.LittleEndian.Uint16(aadLengthBytes)
	aad := make([]byte, addLength)
	_, _ = decoder.Read(aad)
	cipherLengthBytes := make([]byte, 4)
	_, _ = decoder.Read(cipherLengthBytes)
	cipherLength := binary.LittleEndian.Uint32(cipherLengthBytes)
	encrypted := make([]byte, cipherLength+16)
	_, _ = decoder.Read(encrypted)
	hashed := sha256.Sum256(secretKey)
	block, err := aes.NewCipher(hashed[:])
	if err != nil {
		panic(err.Error())
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	plaintext, err := aesgcm.Open(nil, iv, encrypted, aad)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func GetAppContext(header string, secret string) (string, error) {
	zoomApp, err := GetZoomApp()

	if err != nil {
		return "", err
	}

	if len(header) == 0 {
		return "", fmt.Errorf("context header must be a valid string")
	}

	var key string
	if len(secret) > 0 {
		key = secret		
	} else {
		key = zoomApp.ClientSecret;
	}

	// Decode and parse context
	decrypted := ContextDecrypter(header, []byte(key))

	return string(decrypted), nil
}
