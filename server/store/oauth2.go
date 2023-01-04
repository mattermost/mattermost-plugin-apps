// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type OAuth2Store interface {
	CreateState(actingUserID string) (string, error)
	ValidateStateOnce(urlState, actingUserID string) error
	SaveUser(appID apps.AppID, actingUserID string, data []byte) error
	GetUser(appID apps.AppID, actingUserID string) ([]byte, error)
}

type oauth2Store struct {
	*Service
}

var _ OAuth2Store = (*oauth2Store)(nil)

func (s *oauth2Store) CreateState(actingUserID string) (string, error) {
	// fit the max key size of ~50chars
	buf := make([]byte, 15)
	_, _ = rand.Read(buf)
	state := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(buf), actingUserID)
	_, err := s.conf.MattermostAPI().KV.Set(KVOAuth2StatePrefix+state, state, pluginapi.SetExpiry(15*time.Minute))
	if err != nil {
		return "", err
	}
	return state, nil
}

func (s *oauth2Store) ValidateStateOnce(urlState, actingUserID string) error {
	ss := strings.Split(urlState, ".")
	if len(ss) != 2 || ss[1] != actingUserID {
		return utils.ErrForbidden
	}

	storedState := ""
	key := KVOAuth2StatePrefix + urlState
	err := s.conf.MattermostAPI().KV.Get(key, &storedState)
	_ = s.conf.MattermostAPI().KV.Delete(key)
	if err != nil {
		return err
	}
	if storedState != urlState {
		return utils.NewForbiddenError("state mismatch")
	}

	return nil
}

func (s *oauth2Store) SaveUser(appID apps.AppID, actingUserID string, data []byte) error {
	if appID == "" || actingUserID == "" {
		return utils.NewInvalidError("app and user IDs must be provided")
	}

	userkey, err := Hashkey(KVUserPrefix, appID, actingUserID, "", KVUserKey)
	if err != nil {
		return err
	}

	dataEncrypted, err := encrypt([]byte("mykey"), string(data))
	if err != nil {
		return err
	}

	_, err = s.conf.MattermostAPI().KV.Set(userkey, dataEncrypted)
	return err
}

func (s *oauth2Store) GetUser(appID apps.AppID, actingUserID string) ([]byte, error) {
	if appID == "" || actingUserID == "" {
		return nil, utils.NewInvalidError("app and user IDs must be provided")
	}

	userkey, err := Hashkey(KVUserPrefix, appID, actingUserID, "", KVUserKey)
	if err != nil {
		return nil, err
	}

	var data []byte
	if err = s.conf.MattermostAPI().KV.Get(userkey, &data); err != nil {
		return nil, err
	}

	dataDecrypted, err := decrypt([]byte("mykey"), string(data))  // TODO Get key from DB
	if err != nil {
		return nil, err
	}

	return []byte(dataDecrypted), nil
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func encrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "could not create a cipher block, check key")
	}

	msg := pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.Wrap(err, "readFull was unsuccessful, check buffer size")
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
	finalMsg := base64.URLEncoding.EncodeToString(ciphertext)
	return finalMsg, nil
}

func decrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "could not create a cipher block, check key")
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", errors.Wrap(err, "could not decode the message")
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", errors.Wrap(err, "unpad error, check key")
	}

	return string(unpadMsg), nil
}
