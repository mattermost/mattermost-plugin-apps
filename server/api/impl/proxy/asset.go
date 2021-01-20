// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/pkg/errors"
)

func (p *Proxy) Asset(appID api.AppID, assetName string) ([]byte, error) {
	app, err := p.store.LoadApp(appID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't load app - %s", appID)
	}
	for _, asset := range app.Manifest.Assets {
		if asset.Name == assetName {
			if asset.Type == api.S3Asset {
				data, err := p.awsClient.S3FileDownload(asset.Bucket, asset.Key)
				if err != nil {
					return nil, errors.Wrapf(err, "can't download s3 file for appID - %s, assetName - %s from %s/%s", appID, assetName, asset.Bucket, asset.Key)
				}
				return data, nil
			}
			if asset.Type == api.HTTPAsset {
				/* #nosec G107 */
				resp, err := http.Get(asset.URL)
				if err != nil {
					return nil, errors.Wrapf(err, "can't get file for appID - %s, assetName - %s from %s", appID, assetName, asset.URL)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return nil, errors.Errorf("can't get file for appID - %s, assetName - %s from %s. Status - %s", appID, assetName, asset.URL, resp.Status)
				}
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, errors.Wrapf(err, "can't read file for appID - %s, assetName - %s from %s", appID, assetName, asset.URL)
				}
				return bodyBytes, nil
			}
			return nil, errors.New("unknown asset type")
		}
	}
	return nil, errors.New("asset not found")
}
