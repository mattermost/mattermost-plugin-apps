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

func (a *builtinApp) installS3() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "s3",
				Location:    "s3",
				Hint:        "[app ID]",
				Description: "Installs an App from AWS S3, as configured by the system administrator",
				Call:        &installS3Call,
				Form:        appIDForm(installS3Call),
			}
		},

		formf: func(creq apps.CallRequest) (*apps.Form, error) {
			form := appIDForm(installS3Call)
			form.Title = "Install an App from AWS S3"
			form.Fields = append(form.Fields, apps.Field{
				Name:                 fVersion,
				Type:                 apps.FieldTypeDynamicSelect,
				Description:          "select the App's version",
				Label:                fVersion,
				AutocompleteHint:     "app version",
				AutocompletePosition: 2,
			})
			return form, nil
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
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
					return nil, errors.Wrap(err, "failed to retrive the list of app versions on S3, try --url")
				}
				for _, v := range versions {
					options = append(options, apps.SelectOption{
						Value: string(v),
						Label: string(v),
					})
				}
			}

			return options, nil
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			version := creq.GetValue(fVersion, "")
			if version == "" {
				conf, _, log := a.conf.Basic()
				up, err := upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, log)
				if err != nil {
					return apps.NewErrorCallResponse(
						errors.Wrap(err, "failed to initialize AWS access"))
				}

				versions, err := up.ListS3Versions(appID, "")
				if err != nil {
					return apps.NewErrorCallResponse(
						errors.Wrap(err, "failed to retrive the list of app versions on S3"))
				}

				if len(versions) > 0 {
					version = versions[0]
				}
			}

			m, err := a.store.Manifest.GetFromS3(appID, apps.AppVersion(version))
			if err != nil {
				return apps.NewErrorCallResponse(err)
			}

			return a.installCommandSubmit(*m, creq)
		},
	}
}
