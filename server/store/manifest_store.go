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

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type ManifestStore struct {
	*CachedStore[apps.Manifest]
	catalog map[apps.AppID]apps.Manifest
}

func MakeManifestStore(papi plugin.API, mmapi *pluginapi.Client, log utils.Logger) (*ManifestStore, error) {
	cached, err := MakeCachedStore[apps.Manifest](ManifestStoreName, papi, mmapi, log)
	if err != nil {
		return nil, err
	}
	s := &ManifestStore{
		CachedStore: cached,
	}
	return s, nil
}

func (s *ManifestStore) InitCloudCatalog(mmapi *pluginapi.Client, log utils.Logger, conf config.Config, httpOut httpout.Service) error {
	awsClient, err := upaws.MakeClient(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, log)
	if err != nil {
		return errors.Wrap(err, "failed to initialize AWS access")
	}

	bundlePath, err := mmapi.System.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "can't get bundle path for global catalog of apps")
	}
	assetPath := filepath.Join(bundlePath, "assets")
	f, err := os.Open(filepath.Join(assetPath, config.ManifestsFile))
	if err != nil {
		return errors.Wrap(err, "failed to load global catalog of apps")
	}
	defer f.Close()

	catalog := map[apps.AppID]apps.Manifest{}
	manifestLocations := map[apps.AppID]string{}
	err = json.NewDecoder(f).Decode(&manifestLocations)
	if err != nil {
		return errors.Wrap(err, "failed to decode global catalog of apps")
	}

	var data []byte
	for appID, loc := range manifestLocations {
		parts := strings.SplitN(loc, ":", 2)
		switch {
		case len(parts) == 1:
			data, err = getManifestFromS3(awsClient, conf.AWSS3Bucket, appID, apps.AppVersion(parts[0]))
		case len(parts) == 2 && parts[0] == "s3":
			data, err = getManifestFromS3(awsClient, conf.AWSS3Bucket, appID, apps.AppVersion(parts[1]))
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
		catalog[appID] = *m
	}
	s.catalog = catalog
	return nil
}

func (s *ManifestStore) Get(appID apps.AppID) (*apps.Manifest, error) {
	m, ok := s.GetCachedStoreItem(string(appID))
	if ok {
		return &m, nil
	}
	m, ok = s.catalog[appID]
	if ok {
		return &m, nil
	}
	return nil, errors.Wrap(utils.ErrNotFound, string(appID))
}

func (s *ManifestStore) AsMap() map[apps.AppID]apps.Manifest {
	out := map[apps.AppID]apps.Manifest{}
	for appID, m := range s.catalog {
		out[appID] = m
	}
	for id := range s.Index() {
		if m, ok := s.GetCachedStoreItem(id); ok {
			out[apps.AppID(id)] = m
		}
	}
	return out
}

// getManifestFromS3 returns manifest data for an app from the S3
func getManifestFromS3(aws upaws.Client, bucket string, appID apps.AppID, version apps.AppVersion) ([]byte, error) {
	name := upaws.S3ManifestName(appID, version)
	data, err := aws.GetS3(context.Background(), bucket, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download manifest %s", name)
	}
	return data, nil
}

func (s *ManifestStore) Save(r *incoming.Request, m apps.Manifest) error {
	return s.PutCachedStoreItem(r, string(m.AppID), m)
}
