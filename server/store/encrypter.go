package store

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
)

type Encrypter interface {
	Encrypt(message []byte) ([]byte, error)
	Decrypt(message []byte) ([]byte, error)
}

type AESEncrypter struct {
	key []byte
}

var _ Encrypter = (*AESEncrypter)(nil)

func (s *AESEncrypter) unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func (s *AESEncrypter) pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func (s *AESEncrypter) Encrypt(text []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cipher block, check key")
	}

	msg := s.pad(text)
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.Wrap(err, "readFull was unsuccessful, check buffer size")
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
	finalMsg := base64.URLEncoding.EncodeToString(ciphertext)
	return []byte(finalMsg), nil
}

func (s *AESEncrypter) Decrypt(message []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cipher block, check key")
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(string(message))
	if err != nil {
		return nil, errors.Wrap(err, "could not decode the message")
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return nil, errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := s.unpad(msg)
	if err != nil {
		return nil, errors.Wrap(err, "unpad error, check key")
	}

	return unpadMsg, nil
}
