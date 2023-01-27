// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"context" // nolint:gosec
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type ManifestStore struct {
	locallyListed  *CachedStore[apps.Manifest]
	globallyListed map[apps.AppID]apps.Manifest

	aws           upaws.Client
	s3AssetBucket string
}

func MakeManifestStore(papi plugin.API, confService config.Service, httpOut httpout.Service) (*ManifestStore, error) {
	cached, err := MakeCachedStore[apps.Manifest](ManifestStoreName, papi, confService)
	if err != nil {
		return nil, err
	}
	log := confService.NewBaseLogger().With("purpose", "Manifest store")
	conf := confService.Get()
	awsClient, err := upaws.MakeClient(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize AWS access")
	}

	s := &ManifestStore{
		locallyListed: cached,
		aws:           awsClient,
		s3AssetBucket: conf.AWSS3Bucket,
	}

	if conf.MattermostCloudMode {
		bundlePath, err := confService.MattermostAPI().System.GetBundlePath()
		if err != nil {
			return nil, errors.Wrap(err, "can't get bundle path for global catalog of apps")
		}
		assetPath := filepath.Join(bundlePath, "assets")
		f, err := os.Open(filepath.Join(assetPath, config.ManifestsFile))
		if err != nil {
			return nil, errors.Wrap(err, "failed to load global catalog of apps")
		}
		defer f.Close()

		global := map[apps.AppID]apps.Manifest{}
		manifestLocations := map[apps.AppID]string{}
		err = json.NewDecoder(f).Decode(&manifestLocations)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode global catalog of apps")
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
		s.globallyListed = global
	}

	return s, nil
}

func (s *ManifestStore) Get(appID apps.AppID) (*apps.Manifest, error) {
	m, ok := s.locallyListed.Get(string(appID))
	if ok {
		return &m, nil
	}
	m, ok = s.globallyListed[appID]
	if ok {
		return &m, nil
	}
	return nil, errors.Wrap(utils.ErrNotFound, string(appID))
}

func (s *ManifestStore) AsMap() map[apps.AppID]apps.Manifest {
	out := map[apps.AppID]apps.Manifest{}
	for appID, m := range s.globallyListed {
		out[appID] = m
	}
	for id := range s.locallyListed.Index() {
		if m, ok := s.locallyListed.Get(id); ok {
			out[apps.AppID(id)] = m
		}
	}
	return out
}

// getFromS3 returns manifest data for an app from the S3
func (s *ManifestStore) getDataFromS3(appID apps.AppID, version apps.AppVersion) ([]byte, error) {
	name := upaws.S3ManifestName(appID, version)
	data, err := s.aws.GetS3(context.Background(), s.s3AssetBucket, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download manifest %s", name)
	}

	return data, nil
}

// GetFromS3 returns the manifest for an app from the S3
func (s *ManifestStore) GetFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error) {
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

func (s *ManifestStore) Save(r *incoming.Request, m apps.Manifest) error {
	return s.locallyListed.Put(r, string(m.AppID), m)
}

func (s *ManifestStore) PluginClusterEventID() string {
	return s.locallyListed.clusterEventID()
}

func (s *ManifestStore) OnPluginClusterEvent(r *incoming.Request, ev model.PluginClusterEvent) error {
	if ev.Id != s.PluginClusterEventID() {
		return utils.NewInvalidError("unexpected cluster event id: %s", ev.Id)
	}
	return s.locallyListed.processClusterEvent(r, ev.Data)
}
