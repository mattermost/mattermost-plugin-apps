// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) UpdateAppListing(r *incoming.Request, req appclient.UpdateAppListingRequest) (*apps.Manifest, error) {
	if err := r.Check(
		r.RequireSysadminOrPlugin,
	); err != nil {
		return nil, err
	}

	if err := req.Manifest.Validate(); err != nil {
		return nil, utils.NewInvalidError(err, "invalid app manifest in the request")
	}

	if !req.Replace {
		prev, err := p.GetManifest(req.AppID)
		if err != nil && errors.Cause(err) != utils.ErrNotFound {
			return nil, errors.Wrap(err, "failed to load previous listing")
		}
		prevDeploy := apps.Deploy{}
		if prev != nil {
			prevDeploy = prev.Deploy
		}
		req.Manifest.Deploy = mergeDeployData(prevDeploy, req.Manifest.Deploy, req.AddDeploys, req.RemoveDeploys)
	}

	err := p.store.Manifest.StoreLocal(r, req.Manifest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update listed manifest")
	}

	return &req.Manifest, nil
}

func mergeDeployData(prev, d apps.Deploy, addDeploys, removeDeploys apps.DeployTypes) apps.Deploy {
	var result = apps.Deploy{}
	for _, typ := range apps.KnownDeployTypes {
		if !removeDeploys.Contains(typ) {
			result.CopyType(prev, typ)
			if d.Contains(typ) && (prev.Contains(typ) || addDeploys.Contains(typ)) {
				result.CopyType(d, typ)
			}
		}
	}
	return result
}
