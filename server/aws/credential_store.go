// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"
)

const prefixCred = "awscred"

type store struct {
	Mattermost *pluginapi.Client
}

type cred struct {
	KeyID  string
	Secret string
}

type CredentialStoreService interface {
	GetAWSCredential(appID string) (keyID, secret string, err error)
	StoreAWSCredential(appID, keyID, secret string) error
	DeleteAWSCredential(appID string) error
	GetAppList() ([]string, error)
}

func newCredentialStoreService(mm *pluginapi.Client) (CredentialStoreService, error) {
	st := &store{
		Mattermost: mm,
	}
	return st, st.initStore()
}

func (s *store) initStore() error {
	appsKey := getAppsKey()
	var value []string
	err := s.Mattermost.KV.Get(appsKey, &value)
	if err != nil {
		return errors.Wrap(err, "can't get apps list from KV store")
	}
	if value == nil || len(value) == 0 {
		ok, err := s.Mattermost.KV.Set(appsKey, &value)
		if err != nil {
			return errors.Wrapf(err, "can't set app list in KV store")
		}
		if !ok {
			return errors.Errorf("can't set apps list in KV store")
		}
	}
	return nil
}

func (s *store) GetAWSCredential(appID string) (keyID, secret string, err error) {
	key := getKey(appID)
	var value cred
	if err := s.Mattermost.KV.Get(key, value); err != nil {
		return "", "", errors.Wrapf(err, "can't get app %s credential from KV store", appID)
	}
	return value.KeyID, value.Secret, nil
}

func (s *store) StoreAWSCredential(appID, keyID, secret string) error {
	key := getKey(appID)
	value := &cred{
		KeyID:  keyID,
		Secret: secret,
	}
	ok, err := s.Mattermost.KV.Set(key, value)
	if err != nil {
		return errors.Wrapf(err, "can't set app %s credential from KV store", appID)
	}
	if !ok {
		return errors.Errorf("can't set app %s credential from KV store", appID)
	}

	appsKey := getAppsKey()
	var appsList []string
	if err := s.Mattermost.KV.Get(appsKey, &appsList); err != nil {
		return errors.Wrap(err, "can't get apps list from KV store")
	}
	appsList = append(appsList, appID)

	ok, err = s.Mattermost.KV.Set(appsKey, &appsList)
	if err != nil {
		return errors.Wrap(err, "can't set apps list in KV store")
	}
	if !ok {
		return errors.Errorf("can't set apps list in KV store")
	}
	return nil
}

func (s *store) DeleteAWSCredential(appID string) error {
	key := getKey(appID)
	if err := s.Mattermost.KV.Delete(key); err != nil {
		return errors.Errorf("can't delete app %s credential from KV store", appID)
	}

	appsKey := getAppsKey()
	var appsList []string
	if err := s.Mattermost.KV.Get(appsKey, &appsList); err != nil {
		return errors.Wrap(err, "can't get apps list from KV store")
	}
	var newList []string
	for i, app := range appsList {
		if app == appID {
			newList = append(appsList[:i], appsList[i+1:]...)
			break
		}
	}

	ok, err := s.Mattermost.KV.Set(appsKey, &newList)
	if err != nil {
		return errors.Wrap(err, "can't set apps list in KV store")
	}
	if !ok {
		return errors.Errorf("can't set apps list in KV store")
	}
	return nil
}

func (s *store) GetAppList() ([]string, error) {
	key := getAppsKey()
	var apps []string
	if err := s.Mattermost.KV.Get(key, &apps); err != nil {
		return nil, errors.Wrap(err, "can't retreive credential list from KV store")
	}
	return apps, nil
}

func getKey(appID string) string {
	return prefixCred + "_" + appID
}

func getAppsKey() string {
	return prefixCred + "_" + "apps"
}
