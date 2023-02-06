package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type Encrypter interface {
	Encrypt(message string) (string, error)
	Decrypt(message string) ([]byte, error)
}

type AESEncrypter struct {
	key []byte
}

var _ Encrypter = (*AESEncrypter)(nil)

func (s *AESEncrypter) Encrypt(text string) (string, error) {
	byteMsg := []byte(text)
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", errors.Wrap(err, "could not create a cipher block, check key")
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(byteMsg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.Wrap(err, "readFull was unsuccessful, check buffer size")
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *AESEncrypter) Decrypt(message string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return nil, fmt.Errorf("could not base64 decode: %v", err)
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cipher block, check key")
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("blocksize must be multiple of decoded message length")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
