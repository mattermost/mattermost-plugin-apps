// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/sha1" // nolint:gosec
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// manifestStore combines global (aka marketplace) manifests, and locally
// installed ones. The global list is loaded on startup. The local manifests are
// stored in KV store, and the list of their keys is stored in the config, as a
// map of AppID->sha1(manifest).
type manifestStore struct {
	*Store

	mutex sync.RWMutex

	global map[apps.AppID]*apps.Manifest
	local  map[apps.AppID]*apps.Manifest
}

var _ api.ManifestStore = (*manifestStore)(nil)

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

// initGlobal reads in the list of known (i.e. marketplace listed) app
// manifests.
func (s *manifestStore) initGlobal(awscli awsclient.Client, bucket string, manifestsFile io.Reader, assetPath string) error {
	global := map[apps.AppID]*apps.Manifest{}

	manifestLocations := map[apps.AppID]string{}
	err := json.NewDecoder(manifestsFile).Decode(&manifestLocations)
	if err != nil {
		return err
	}

	var data []byte
	for appID, loc := range manifestLocations {
		parts := strings.SplitN(loc, ":", 2)
		switch {
		case len(parts) == 1:
			data, err = s.getFromS3(awscli, bucket, appID, apps.AppVersion(parts[0]))
		case len(parts) == 2 && parts[0] == "s3":
			data, err = s.getFromS3(awscli, bucket, appID, apps.AppVersion(parts[1]))
		case len(parts) == 2 && parts[0] == "file":
			data, err = ioutil.ReadFile(filepath.Join(assetPath, parts[1]))
		case len(parts) == 2 && (parts[0] == "http" || parts[0] == "https"):
			data, err = httputils.GetFromURL(loc)
		default:
			s.mm.Log.Error("failed to load global manifest",
				"err", fmt.Sprintf("%s is invalid", loc),
				"app_id", appID)
			continue
		}
		if err != nil {
			s.mm.Log.Error("failed to load global manifest",
				"err", err.Error(),
				"app_id", appID,
				"loc", loc)
			continue
		}

		var m *apps.Manifest
		m, err = apps.ManifestFromJSON(data)
		if err != nil {
			s.mm.Log.Error("failed to load global manifest",
				"err", err.Error(),
				"app_id", appID,
				"loc", loc)
			continue
		}
		if m.AppID != appID {
			s.mm.Log.Error("failed to load global manifest",
				"err", fmt.Sprintf("mismatched app ids while getting manifest %s != %s", m.AppID, appID),
				"app_id", appID,
				"loc", loc)
			continue
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
	err = m.IsValid()
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *manifestStore) Configure(conf api.Config) {
	updatedLocal := map[apps.AppID]*apps.Manifest{}

	for id, key := range conf.LocalManifests {
		var m *apps.Manifest
		err := s.mm.KV.Get(api.PrefixLocalManifest+key, &m)
		switch {
		case err != nil:
			s.mm.Log.Error(
				fmt.Sprintf("failed to load local manifest for %s: %s", id, err.Error()))

		case m == nil:
			s.mm.Log.Error(
				fmt.Sprintf("failed to load local manifest for %s: not found", id))

		default:
			updatedLocal[apps.AppID(id)] = m
		}
	}

	s.mutex.Lock()
	s.local = updatedLocal
	s.mutex.Unlock()
}

func (s *manifestStore) Get(appID apps.AppID) (*apps.Manifest, error) {
	s.mutex.RLock()
	local := s.local
	global := s.global
	s.mutex.RUnlock()

	m, ok := local[appID]
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

	err = s.mm.KV.Delete(api.PrefixLocalManifest + prevSHA)
	if err != nil {
		s.mm.Log.Warn("failed to delete previous Manifest KV value", "err", err.Error())
	}
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

// getFromS3 returns a manifest file for an app from the S3
func (s *manifestStore) getFromS3(awscli awsclient.Client, bucket string, appID apps.AppID, version apps.AppVersion) ([]byte, error) {
	name := apps.ManifestS3Name(appID, version)
	data, err := awscli.GetS3(bucket, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download manifest %s", name)
	}
	return data, nil
}
