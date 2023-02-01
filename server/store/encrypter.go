package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
)

type Encrypter interface {
	Encrypt(message []byte) ([]byte, error)
	Decrypt(message []byte) ([]byte, error)
}

type AESEncrypter struct {
	key string
}

var _ Encrypter = (*AESEncrypter)(nil)

func (s *AESEncrypter) Encrypt(text []byte) ([]byte, error) {
	key, err := hex.DecodeString(s.key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cipher block, check key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cipher block, check key")
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.Wrap(err, "readFull was unsuccessful, check buffer size")
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return ciphertext, nil
}

func (s *AESEncrypter) Decrypt(message []byte) ([]byte, error) {
	key, err := hex.DecodeString(s.key)
	ciphertext, _ := hex.DecodeString(string(message))
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cipher block, check key")
	}

	block, err := aes.NewCipher(key)
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

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
