package domain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"drawwwingame/domain/valobj"
	"encoding/hex"
	"io"
	"strings"
)

const (
	ENCRYPT_KEY_HEX = "22b0720abd3a6f7a954037fcc5e5098a"
)

func Encrypt(str string) (string, error) { // not contain "~"
	key, err := hex.DecodeString(ENCRYPT_KEY_HEX)
	if err != nil {
		Log(err)
		return "", ErrorInternal
	}
	if strings.Contains(str, "~") {
		LogStringf("str contain ~")
		return "", ErrorString
	}
	str += "~"
	if len(str)%aes.BlockSize != 0 {
		add_length := aes.BlockSize - len(str)%aes.BlockSize
		str += valobj.NewAlphanumStringRandom(add_length).ToString()
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		Log(err)
		return "", ErrorString
	}
	plain_text := []byte(str)
	cipher_text := make([]byte, aes.BlockSize+len(plain_text))
	iv := cipher_text[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		Log(err)
		return "", ErrorString
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipher_text[aes.BlockSize:], plain_text)
	return hex.EncodeToString(cipher_text), nil
}

func Decrypt(str string) (string, error) {
	key, err := hex.DecodeString(ENCRYPT_KEY_HEX)
	if err != nil {
		Log(err)
		return "", ErrorInternal
	}
	cipher_str, err := hex.DecodeString(str)
	if err != nil {
		Log(err)
		return "", ErrorInternal
	}
	cipher_text := []byte(cipher_str)

	block, err := aes.NewCipher(key)
	if err != nil {
		Log(err)
		return "", ErrorString
	}
	if len(cipher_text) < aes.BlockSize {
		LogStringf("length too short")
		return "", ErrorString
	}
	iv := cipher_text[:aes.BlockSize]
	cipher_text = cipher_text[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipher_text, cipher_text)
	s := string(cipher_text)
	index := strings.Index(s, "~")
	return s[:index], nil
}
