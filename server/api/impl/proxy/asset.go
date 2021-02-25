// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
)

func (p *Proxy) GetAsset(appID apps.AppID, assetName string) (*http.Response, error) {
	app, err := p.store.App().Get(appID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't load app - %s", appID)
	}
	errorMessage := fmt.Sprintf("can't download %s for appID - %s, assetName - %s", app.Manifest.Type, appID, assetName)
	switch app.Manifest.Type {
	case apps.AppTypeAWSLambda:
		key := aws.GetAssetFileKey(app.AppID, app.Manifest.Version, assetName)
		data, err := p.awsClient.S3AssetDownload(key)
		if err != nil {
			return nil, errors.Wrapf(err, errorMessage)
		}
		resp := &http.Response{}
		resp.StatusCode = http.StatusOK
		resp.Body = ioutil.NopCloser(bytes.NewReader(data))
		return resp, nil
	case apps.AppTypeHTTP:
		url := fmt.Sprintf("%s/static/%s", app.Manifest.HTTPRootURL, assetName)
		/* #nosec G107 */
		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.Wrapf(err, "%s, url - %s", errorMessage, url)
		}
		return resp, nil
	case apps.AppTypeBuiltin:
		return nil, errors.New("assets are not supported yet for builtin apps")
	default:
		return nil, errors.New("asset not found, unknown app type")
	}
}
