// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) GetAsset(appID apps.AppID, assetName string) ([]byte, error) {
	app, err := p.store.App().Get(appID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't load app - %s", appID)
	}
	for _, asset := range app.Manifest.Assets {
		if asset.Name != assetName {
			continue
		}
		errorMessage := fmt.Sprintf("can't download %s for appID - %s, assetName - %s from", asset.Type, appID, assetName)
		switch asset.Type {
		case apps.S3Asset:
			data, err := p.awsClient.S3FileDownload(asset.Bucket, asset.Key)
			if err != nil {
				return nil, errors.Wrapf(err, "%s %s/%s", errorMessage, asset.Bucket, asset.Key)
			}
			return data, nil
		case apps.HTTPAsset:
			/* #nosec G107 */
			resp, err := http.Get(asset.URL)
			if err != nil {
				return nil, errors.Wrapf(err, "%s %s", errorMessage, asset.URL)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return nil, errors.Errorf("%s %s. Status - %s", errorMessage, asset.URL, resp.Status)
			}
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.Wrapf(err, "%s %s", errorMessage, asset.URL)
			}
			return bodyBytes, nil
		}
		return nil, errors.New("unknown asset type")
	}
	return nil, errors.New("asset not found")
}
