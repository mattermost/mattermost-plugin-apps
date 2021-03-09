// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

func (p *Proxy) GetAsset(appID apps.AppID, assetName string) (io.ReadCloser, int, error) {
	app, err := p.store.App().Get(appID)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "can't load app - %s", appID)
	}

	errorMessage := fmt.Sprintf("can't download %s for appID - %s, assetName - %s", app.Manifest.Type, appID, assetName)
	switch app.Manifest.Type {
	case apps.AppTypeAWSLambda:
		key := awsclient.GenerateAssetS3Name(app.AppID, app.Manifest.Version, assetName)
		data, err := p.aws.Client().GetS3(p.s3AssetBucket, key)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrapf(err, errorMessage)
		}
		return ioutil.NopCloser(bytes.NewReader(data)), http.StatusOK, nil

	case apps.AppTypeHTTP:
		url := fmt.Sprintf("%s/static/%s", app.Manifest.HTTPRootURL, assetName)
		/* #nosec G107 */
		resp, err := http.Get(url) // nolint:bodyclose
		if err != nil {
			return nil, http.StatusBadGateway, errors.Wrapf(err, "%s, url - %s", errorMessage, url)
		}
		return resp.Body, resp.StatusCode, nil

	case apps.AppTypeBuiltin:
		return nil, http.StatusBadRequest, errors.New("assets are not supported yet for builtin apps")

	default:
		return nil, http.StatusBadRequest, errors.New("asset not found, unknown app type")
	}
}
