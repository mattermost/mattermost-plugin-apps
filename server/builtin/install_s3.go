// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/pkg/errors"
)

var installS3Call = apps.Call{
	Path: pInstallS3,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll, // ensure sysadmin
	},
}

func (a *builtinApp) installS3Form(creq apps.CallRequest) apps.CallResponse {
	resp := apps.CallResponse{
		Form: appIDForm(installS3Call),
	}
	resp.Form.Title = "Install an App from AWS S3"
	resp.Form.Fields = append(resp.Form.Fields, apps.Field{
		Name:                 fVersion,
		Type:                 apps.FieldTypeDynamicSelect,
		Description:          "select the App's version",
		Label:                fVersion,
		AutocompleteHint:     "app version",
		AutocompletePosition: 2,
	})
	return resp
}

func (a *builtinApp) installS3Lookup(creq apps.CallRequest) ([]apps.SelectOption, error) {
	if creq.SelectedField != fAppID && creq.SelectedField != fVersion {
		return nil, errors.Errorf("unknown field %q", creq.SelectedField)
	}

	conf, _, log := a.conf.Basic()
	up, err := upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize AWS access")
	}

	var options []apps.SelectOption
	switch creq.SelectedField {
	case fAppID:
		appIDs, err := up.ListS3Apps(creq.Query)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrive the list of apps, try --url")
		}
		for _, appID := range appIDs {
			options = append(options, apps.SelectOption{
				Value: string(appID),
				Label: string(appID),
			})
		}

	case fVersion:
		id := creq.GetValue(fAppID, "")
		versions, err := up.ListS3Versions(apps.AppID(id), creq.Query)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrive the list of apps, try --url")
		}
		for _, v := range versions {
			options = append(options, apps.SelectOption{
				Value: string(v),
				Label: string(v),
			})
		}
	}

	return options, nil
}

func (a *builtinApp) installS3Submit(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	version := apps.AppVersion(creq.GetValue(fVersion, ""))
	m, err := a.store.Manifest.GetFromS3(appID, version)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return a.installCommandSubmit(*m, creq)
}
