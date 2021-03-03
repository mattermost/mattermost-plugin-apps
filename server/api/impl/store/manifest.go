// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/sha1" // nolint:gosec
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type manifestStore struct {
	*Store

	mutex sync.RWMutex

	global  map[apps.AppID]*apps.Manifest
	local   map[apps.AppID]*apps.Manifest
	builtin map[apps.AppID]*apps.Manifest
}

var _ api.ManifestStore = (*manifestStore)(nil)

func (s *manifestStore) InitBuiltin(manifests ...*apps.Manifest) {
	s.mutex.Lock()
	if s.builtin == nil {
		s.builtin = map[apps.AppID]*apps.Manifest{}
	}
	for _, m := range manifests {
		s.builtin[m.AppID] = m
	}
	s.mutex.Unlock()
}

func (s *manifestStore) InitGlobal(awscli awsclient.Client, bucket string) error {
	bundlePath, err := s.mm.System.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "can't get bundle path")
	}
	assetPath := filepath.Join(bundlePath, "assets")
	f, err := os.Open(filepath.Join(assetPath, api.ManifestsFile))
	if err != nil {
		return errors.Wrap(err, "failed to load global list of available apps")
	}
	defer f.Close()

	return s.initGlobal(awscli, bucket, f, assetPath)
}

func (s *manifestStore) initGlobal(awscli awsclient.Client, bucket string, manifestsFile io.Reader, assetPath string) error {
	global := map[apps.AppID]*apps.Manifest{}

	// Read in the marketplace-listed manifests from S3, as per versions
	// indicated in apps.json. apps.json file contains a map of AppID->manifest
	// S3 filename (the bucket comes from the config)
	manifestLocations := apps.AppVersionMap{}
	err := json.NewDecoder(manifestsFile).Decode(&manifestLocations)
	if err != nil {
		return err
	}

	var data []byte
	for appID, loc := range manifestLocations {
		parts := strings.SplitN(string(loc), ":", 2)
		switch {
		case len(parts) == 1:
			data, err = s.getFromS3(awscli, bucket, appID, apps.AppVersion(parts[0]))
		case len(parts) == 2 && parts[0] == "s3":
			data, err = s.getFromS3(awscli, bucket, appID, apps.AppVersion(parts[1]))
		case len(parts) == 2 && parts[0] == "file":
			data, err = ioutil.ReadFile(filepath.Join(assetPath, parts[1]))
		case len(parts) == 2 && (parts[0] == "http" || parts[0] == "https"):
			data, err = httputils.GetFromURL(string(loc))
		default:
			return errors.Errorf("failed to load global manifest for %s: %s is invalid", string(appID), loc)
		}
		if err != nil {
			return errors.Wrapf(err, "failed to load global manifest for %s", string(appID))
		}

		var m *apps.Manifest
		m, err = DecodeManifest(data)
		if err != nil {
			return err
		}
		if m.AppID != appID {
			return errors.Errorf("mismatched app ids while getting manifest %s != %s", m.AppID, appID)
		}
		global[appID] = m
	}

	s.mutex.Lock()
	s.global = global
	s.mutex.Unlock()

	return nil
}

func DecodeManifest(data []byte) (*apps.Manifest, error) {
	var m apps.Manifest
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	err = validateManifest(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *manifestStore) Configure(conf api.Config) error {
	updatedLocal := map[apps.AppID]*apps.Manifest{}

	for id, key := range conf.LocalManifests {
		var m *apps.Manifest
		err := s.mm.KV.Get(api.PrefixLocalManifest+key, &m)
		if err != nil {
			s.mm.Log.Error(
				fmt.Sprintf("failed to load local manifest for %s: %s", id, err.Error()))
		}
		if m == nil {
			s.mm.Log.Error(
				fmt.Sprintf("failed to load local manifest for %s: not found", id))
		}

		updatedLocal[apps.AppID(id)] = m
	}

	s.mutex.Lock()
	s.local = updatedLocal
	s.mutex.Unlock()

	return nil
}

func (s *manifestStore) Get(appID apps.AppID) (*apps.Manifest, error) {
	s.mutex.RLock()
	builtin := s.builtin
	local := s.local
	global := s.global
	s.mutex.RUnlock()

	m, ok := builtin[appID]
	if ok {
		return m, nil
	}
	m, ok = local[appID]
	if ok {
		return m, nil
	}
	m, ok = global[appID]
	if ok {
		return m, nil
	}
	return nil, utils.ErrNotFound
}

func (s *manifestStore) AsMap() map[apps.AppID]*apps.Manifest {
	s.mutex.RLock()
	builtin := s.builtin
	local := s.local
	global := s.global
	s.mutex.RUnlock()

	out := map[apps.AppID]*apps.Manifest{}
	for id, m := range global {
		out[id] = m
	}
	for id, m := range local {
		out[id] = m
	}
	for id, m := range builtin {
		out[id] = m
	}
	return out
}

func (s *manifestStore) StoreLocal(m *apps.Manifest) error {
	conf := s.conf.GetConfig()
	prevSHA := conf.LocalManifests[string(m.AppID)]

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	sha := fmt.Sprintf("%x", sha1.Sum(data)) // nolint:gosec
	if sha == prevSHA {
		return nil
	}

	_, err = s.mm.KV.Set(api.PrefixLocalManifest+sha, m)
	if err != nil {
		return err
	}

	s.mutex.RLock()
	local := s.local
	s.mutex.RUnlock()
	updatedLocal := map[apps.AppID]*apps.Manifest{}
	for k, v := range local {
		if k != m.AppID {
			updatedLocal[k] = v
		}
	}
	updatedLocal[m.AppID] = m
	s.mutex.Lock()
	s.local = updatedLocal
	s.mutex.Unlock()

	updated := map[string]string{}
	for k, v := range conf.LocalManifests {
		updated[k] = v
	}
	updated[string(m.AppID)] = sha
	sc := *conf.StoredConfig
	sc.LocalManifests = updated
	err = s.conf.StoreConfig(&sc)
	if err != nil {
		return err
	}

	_ = s.mm.KV.Delete(api.PrefixLocalManifest + prevSHA)
	return nil
}

func (s *manifestStore) DeleteLocal(appID apps.AppID) error {
	conf := s.conf.GetConfig()
	sha := conf.LocalManifests[string(appID)]

	err := s.mm.KV.Delete(api.PrefixLocalManifest + sha)
	if err != nil {
		return err
	}

	s.mutex.RLock()
	local := s.local
	s.mutex.RUnlock()
	updatedLocal := map[apps.AppID]*apps.Manifest{}
	for k, v := range local {
		if k != appID {
			updatedLocal[k] = v
		}
	}
	s.mutex.Lock()
	s.local = updatedLocal
	s.mutex.Unlock()

	updated := map[string]string{}
	for k, v := range conf.LocalManifests {
		updated[k] = v
	}
	delete(updated, string(appID))
	sc := *conf.StoredConfig
	sc.LocalManifests = updated

	return s.conf.StoreConfig(&sc)
}

// GetManifest returns a manifest file for an app from the S3
func (s *manifestStore) getFromS3(awscli awsclient.Client, bucket string, appID apps.AppID, version apps.AppVersion) ([]byte, error) {
	name := fmt.Sprintf("manifest_%s_%s", appID, version)
	data, err := awscli.GetS3(bucket, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download manifest %s", name)
	}
	return data, nil
}

func validateManifest(m *apps.Manifest) error {
	if m.AppID == "" {
		return errors.New("empty AppID")
	}
	if m.Type == "" {
		return errors.New("app_type is empty, must be specified, e.g. `aws_lamda`")
	}
	if !m.Type.IsValid() {
		return errors.Errorf("invalid type: %s", m.Type)
	}

	if m.Type == apps.AppTypeHTTP {
		_, err := url.Parse(m.HTTPRootURL)
		if err != nil {
			return errors.Wrapf(err, "invalid manifest URL %q", m.HTTPRootURL)
		}
	}
	return nil
}