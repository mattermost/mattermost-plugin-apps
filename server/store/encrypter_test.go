package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func TestEncrypterEncode(t *testing.T) {
	key, err := config.GenerateEncryptionKey()
	assert.NoError(t, err)

	for _, tc := range []struct {
		name          string
		message       []byte
		expectedError string
		key           string
	}{
		{
			name:          "The key is not valid",
			message:       nil,
			expectedError: "could not create a cipher block, check key: crypto/aes: invalid key size 0",
			key:           "",
		},
		{
			name:          "The message is encrypted with a generated valid key",
			message:       []byte(`{"Test1":"test-1","Test2":"test-2"}`),
			expectedError: "",
			key:           key,
		},
		{
			name:          "The message is encrypted",
			message:       []byte(`{"Test1":"test-1","Test2":"test-2"}`),
			expectedError: "",
			key:           "6368616e676520746869732070617373",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := &AESEncrypter{key: tc.key}

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
		messageEncrypted []byte
		expected         string
		expectedError    string
		key              string
	}{
		{
			name:             "The key is not valid",
			messageEncrypted: []byte(""),
			expected:         "",
			expectedError:    "could not create a cipher block, check key: crypto/aes: invalid key size 0",
			key:              "",
		},
		{
			name:             "The key is valid and the message decoded",
			messageEncrypted: []byte("67ef87bec4a7d5f8f6e889241788c666af162ab02be3ef6e79a4a514c398536a6f543d400374443e4882d52c2c38c9f06a9cd7"),
			expected:         `{"Test1":"test-1","Test2":"test-2"}`,
			expectedError:    "",
			key:              "6368616e676520746869732070617373",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := &AESEncrypter{key: tc.key}

			decryptedMessage, err := encrypter.Decrypt(tc.messageEncrypted)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.Equal(t, tc.expected, string(decryptedMessage))
			}
		})
	}
}
