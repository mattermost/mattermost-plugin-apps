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
		message       string
		isBase64      bool
		expectedError string
		key           []byte
	}{
		{
			name:          "The key is not valid",
			message:       "",
			isBase64:      false,
			expectedError: "could not create a cipher block, check key: crypto/aes: invalid key size 0",
			key:           []byte(""),
		},
		{
			name:          "The message is encrypted with a generated valid key",
			message:       `{"Test1":"test-1","Test2":"test-2"}`,
			isBase64:      true,
			expectedError: "",
			key:           key,
		},
		{
			name:          "The message is encrypted",
			message:       `{"Test1":"test-1","Test2":"test-2"}`,
			isBase64:      true,
			expectedError: "",
			key:           []byte("6368616e676520746869732070617373"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const regexBase64 = "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{4})$"

			encrypter := &AESEncrypter{key: tc.key}

			encryptedItem, err := encrypter.Encrypt(tc.message)

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.Regexp(t, regexBase64, string(encryptedItem))
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
			expectedError:    "could not create a cipher block, check key: crypto/aes: invalid key size 0",
			key:              []byte(""),
		},
		{
			name:             "The key is valid but the message is not base64 encoded",
			messageEncrypted: "67ef87bec4a7d5f8f6e889241788c666af162ab02be3ef6e79a4a514c398536a6f543d400374443e4882d52c2c38c9f06a9cd7",
			expected:		  "",
			expectedError:    "could not base64 decode: illegal base64 data at input byte 100",
			key:              []byte("6368616e676520746869732070617373"),
		},
		{
			name:             "The key is valid and the message decoded",
			messageEncrypted: "5MJMe6KixZJfxnRw2RYoRoGSW3W2GQA1+XKNf4gM1jyKQluH5zqWpmsjcP/kclwCyNsU",
			expected:         `{"Test1":"test-1","Test2":"test-2"}`,
			expectedError:    "",
			key:              []byte("6368616e676520746869732070617373"),
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
