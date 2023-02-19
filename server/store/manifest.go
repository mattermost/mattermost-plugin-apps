// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"context"
	"crypto/sha1" // nolint:gosec
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type ManifestStore interface {
	config.Configurable

	StoreLocal(*incoming.Request, apps.Manifest) error
	Get(apps.AppID) (*apps.Manifest, error)
	GetFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error)
	AsMap() map[apps.AppID]apps.Manifest
	DeleteLocal(*incoming.Request, apps.AppID) error
}

// manifestStore combines global (aka marketplace) manifests, and locally
// installed ones. The global list is loaded on startup. The local manifests are
// stored in KV store, and the list of their keys is stored in the config, as a
// map of AppID->sha1(manifest).
type manifestStore struct {
	*Service

	// mutex guards local, the pointer to the map of locally-installed
	// manifests.
	mutex sync.RWMutex

	global map[apps.AppID]apps.Manifest
	local  map[apps.AppID]apps.Manifest

	aws           upaws.Client
	s3AssetBucket string
}

var _ ManifestStore = (*manifestStore)(nil)

func (s *Service) makeManifestStore() (*manifestStore, error) {
	conf := s.conf.Get()
	log := s.conf.NewBaseLogger().With("purpose", "Manifest store")
	awsClient, err := upaws.MakeClient(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize AWS access")
	}

	mstore := &manifestStore{
		Service:       s,
		aws:           awsClient,
		s3AssetBucket: conf.AWSS3Bucket,
	}
	if err = mstore.Configure(log); err != nil {
		return nil, errors.Wrap(err, "failed to configure")
	}

	if conf.MattermostCloudMode {
		err = mstore.InitGlobal(s.httpOut, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize the global manifest list from marketplace")
		}
	}
	return mstore, nil
}

// InitGlobal reads in the list of known (i.e. marketplace listed) app
// manifests.
func (s *manifestStore) InitGlobal(httpOut httpout.Service, log utils.Logger) error {
	conf := s.conf.Get()

	bundlePath, err := s.conf.API().Mattermost.System.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "can't get bundle path")
	}
	assetPath := filepath.Join(bundlePath, "assets")
	f, err := os.Open(filepath.Join(assetPath, config.ManifestsFile))
	if err != nil {
		return errors.Wrap(err, "failed to load global list of available apps")
	}
	defer f.Close()

	global := map[apps.AppID]apps.Manifest{}
	manifestLocations := map[apps.AppID]string{}
	err = json.NewDecoder(f).Decode(&manifestLocations)
	if err != nil {
		return err
	}

	var data []byte
	for appID, loc := range manifestLocations {
		parts := strings.SplitN(loc, ":", 2)
		switch {
		case len(parts) == 1:
			data, err = s.getDataFromS3(appID, apps.AppVersion(parts[0]))
		case len(parts) == 2 && parts[0] == "s3":
			data, err = s.getDataFromS3(appID, apps.AppVersion(parts[1]))
		case len(parts) == 2 && parts[0] == "file":
			data, err = os.ReadFile(filepath.Join(assetPath, parts[1]))
		case len(parts) == 2 && (parts[0] == "http" || parts[0] == "https"):
			data, err = httpOut.GetFromURL(loc, conf.DeveloperMode, apps.MaxManifestSize)
		default:
			log.WithError(err).Errorw("failed to load global manifest",
				"app_id", appID)
			continue
		}
		if err != nil {
			log.WithError(err).Errorw("failed to load global manifest",
				"app_id", appID,
				"loc", loc)
			continue
		}

		m, err := apps.DecodeCompatibleManifest(data)
		if err != nil {
			log.WithError(err).Errorw("failed to load global manifest",
				"app_id", appID,
				"loc", loc)
			continue
		}
		if m.AppID != appID {
			err = errors.Errorf("mismatched app ids while getting manifest %s != %s", m.AppID, appID)
			log.WithError(err).Errorw("failed to load global manifest",
				"app_id", appID,
				"loc", loc)
			continue
		}
		global[appID] = *m
	}

	s.mutex.Lock()
	s.global = global
	s.mutex.Unlock()

	return nil
}

func (s *manifestStore) Configure(log utils.Logger) error {
	updatedLocal := map[apps.AppID]apps.Manifest{}

	for id, key := range s.conf.Get().LocalManifests {
		log = log.With("app_id", id)

		var data []byte
		err := s.conf.API().Mattermost.KV.Get(KVLocalManifestPrefix+key, &data)
		if err != nil {
			log.WithError(err).Errorw("Failed to get local manifest from KV")
			continue
		}

		if len(data) == 0 {
			err = utils.NewNotFoundError(KVLocalManifestPrefix + key)
			log.WithError(err).Errorw("Failed to load local manifest")
			continue
		}

		m, err := apps.DecodeCompatibleManifest(data)
		if err != nil {
			log.WithError(err).Errorw("Failed to decode local manifest")
			continue
		}
		updatedLocal[apps.AppID(id)] = *m
	}

	s.mutex.Lock()
	s.local = updatedLocal
	s.mutex.Unlock()
	return nil
}

func (s *manifestStore) Get(appID apps.AppID) (*apps.Manifest, error) {
	s.mutex.RLock()
	local := s.local
	global := s.global
	s.mutex.RUnlock()

	m, ok := local[appID]
	if ok {
		return &m, nil
	}
	m, ok = global[appID]
	if ok {
		return &m, nil
	}
	return nil, errors.Wrap(utils.ErrNotFound, string(appID))
}

func (s *manifestStore) AsMap() map[apps.AppID]apps.Manifest {
	s.mutex.RLock()
	local := s.local
	global := s.global
	s.mutex.RUnlock()

	out := map[apps.AppID]apps.Manifest{}
	for id, m := range global {
		out[id] = m
	}
	for id, m := range local {
		out[id] = m
	}
	return out
}

func (s *manifestStore) StoreLocal(r *incoming.Request, m apps.Manifest) error {
	conf := s.conf.Get()
	mm := s.conf.API().Mattermost
	prevSHA := conf.LocalManifests[string(m.AppID)]

	m.SchemaVersion = conf.PluginManifest.Version
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	sha := fmt.Sprintf("%x", sha1.Sum(data)) // nolint:gosec
	_, err = mm.KV.Set(KVLocalManifestPrefix+sha, m)
	if err != nil {
		return err
	}

	s.mutex.RLock()
	local := s.local
	s.mutex.RUnlock()
	updatedLocal := map[apps.AppID]apps.Manifest{}
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
	sc := conf.StoredConfig
	sc.LocalManifests = updated
	err = s.conf.StoreConfig(sc, r.Log)
	if err != nil {
		return err
	}

	if sha != prevSHA {
		err = mm.KV.Delete(KVLocalManifestPrefix + prevSHA)
		if err != nil {
			r.Log.WithError(err).Warnf("Failed to delete previous Manifest KV value")
		}
	}
	return nil
}

func (s *manifestStore) DeleteLocal(r *incoming.Request, appID apps.AppID) error {
	conf := s.conf.Get()
	sha := conf.LocalManifests[string(appID)]

	err := s.conf.API().Mattermost.KV.Delete(KVLocalManifestPrefix + sha)
	if err != nil {
		return err
	}

	s.mutex.RLock()
	local := s.local
	s.mutex.RUnlock()
	updatedLocal := map[apps.AppID]apps.Manifest{}
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
	sc := conf.StoredConfig
	sc.LocalManifests = updated

	return s.conf.StoreConfig(sc, r.Log)
}

// getFromS3 returns manifest data for an app from the S3
func (s *manifestStore) getDataFromS3(appID apps.AppID, version apps.AppVersion) ([]byte, error) {
	name := upaws.S3ManifestName(appID, version)
	data, err := s.aws.GetS3(context.Background(), s.s3AssetBucket, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download manifest %s", name)
	}

	return data, nil
}

// GetFromS3 returns the manifest for an app from the S3
func (s *manifestStore) GetFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error) {
	data, err := s.getDataFromS3(appID, version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manifest data")
	}

	m, err := apps.DecodeCompatibleManifest(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal manifest data")
	}

	if m.AppID != appID {
		return nil, errors.New("mismatched app ID")
	}

	if m.Version != version {
		return nil, errors.New("mismatched app version")
	}

	return m, nil
}
