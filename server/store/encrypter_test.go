package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypterEncode(t *testing.T) {

	for _, tc := range []struct {
		name          string
		message       string
		expectedError string
		key           []byte
	}{
		{
			name:          "The key is not valid",
			message:       "",
			expectedError: "could not create a cipher block, check key: crypto/aes: invalid key size 7",
			key:           []byte("invalid"),
		},
		{
			name:          "The message is encrypted",
			message:       `{"Test1":"test-1","Test2":"test-2"}`,
			expectedError: "",
			key:           []byte("asuperstrong32bitpasswordgohere!"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := &StoreEncrypter{key: tc.key}

			encryptedItem, err := encrypter.Encrypt(tc.message)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NotEmpty(t, encryptedItem)
			}
		})
	}
}

func TestEncrypterDecrypt(t *testing.T) {
	for _, tc := range []struct {
		name             string
		messageEncrypted string
		expected         string
		expectedError    string
		key              []byte
	}{
		{
			name:             "The key is not valid",
			messageEncrypted: "",
			expected:         "",
			expectedError:    "could not create a cipher block, check key: crypto/aes: invalid key size 7",
			key:              []byte("invalid"),
		},
		{
			name:             "The key is valid but the message stored is invalid",
			messageEncrypted: "AAAAAAAAAAAA",
			expected:         "",
			expectedError:    "blocksize must be multiple of decoded message length",
			key:              []byte("asuperstrong32bitpasswordgohere!"),
		},
		{
			name:             "The key is valid and the message decoded",
			messageEncrypted: "qrZ7JgEW2hi37toQsTorIZSqLv4xRDyHfQulLziP3UonAP77idbimFk9dRObgDgOlJj8E9rrFna0ESpSFFj4UQ==",
			expected:         `{"Test1":"test-1","Test2":"test-2"}`,
			expectedError:    "",
			key:              []byte("asuperstrong32bitpasswordgohere!"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := &StoreEncrypter{key: tc.key}

			decryptedMessage, err := encrypter.Decrypt(tc.messageEncrypted)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.Equal(t, tc.expected, decryptedMessage)
			}

		})
	}
}
