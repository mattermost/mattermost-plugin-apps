package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypterEncode(t *testing.T) {
	for _, tc := range []struct {
		name          string
		message       string
		expected      string
		expectedError string
		key           []byte
	}{
		{
			name:          "The key is not valid",
			message:       "",
			expected:      "",
			expectedError: "could not create a cipher block, check key: crypto/aes: invalid key size 7",
			key:           []byte("invalid"),
		},
		// {
		// 	name:          "The key is valid but we couldn't decrypt the message",
		// 	message:       "",
		// 	expected:      "",
		// 	expectedError: "",
		// 	key:           []byte("invalid"),
		// },
		// {
		// 	name:          "",
		// 	message:       "",
		// 	expected:      "",
		// 	expectedError: "",
		// 	key:           []byte("invalid"),
		// },
	} {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := &StoreEncrypter{key: tc.key}

			encryptedItem, err := encrypter.Encrypt(tc.message)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			}

			assert.Equal(t, tc.expected, encryptedItem)
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
		// {
		// 	name:             "The key is valid but we couldn't decrypt the message",
		// 	messageEncrypted: "",
		// 	expected:         "",
		// 	expectedError:    "",
		// 	key:              []byte("invalid"),
		// },
		// {
		// 	name:             "",
		// 	messageEncrypted: "",
		// 	expected:         "",
		// 	expectedError:    "",
		// 	key:              []byte("invalid"),
		// },
	} {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := &StoreEncrypter{key: tc.key}

			decryptedMessage, err := encrypter.Decrypt(tc.messageEncrypted)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			}

			assert.Equal(t, tc.expected, decryptedMessage)
		})
	}
}
