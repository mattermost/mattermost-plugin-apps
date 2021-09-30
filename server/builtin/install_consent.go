// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func (a *builtinApp) installConsent() handler {
	return handler{
		requireSysadmin: true,

		formf: func(creq apps.CallRequest) (*apps.Form, error) {
			id, ok := creq.State.(string)
			if !ok {
				return nil, errors.New("no app ID in state, don't know what to install")
			}
			appID := apps.AppID(id)

			m, err := a.store.Manifest.Get(appID)
			if err != nil {
				return nil, err
			}

			return a.newInstallConsentForm(*m, creq), nil
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			secret := creq.GetValue(fSecret, "")
			consent := creq.BoolValue(fConsent)
			id, ok := creq.State.(string)
			if !ok {
				return apps.NewErrorCallResponse(
					errors.New("no app ID in state, don't know what to install"))
			}
			appID := apps.AppID(id)

			m, err := a.store.Manifest.Get(appID)
			if err != nil {
				return apps.NewErrorCallResponse(errors.Wrap(err, "failed to load App manifest"))
			}
			if !consent && len(m.RequestedLocations)+len(m.RequestedPermissions) > 0 {
				return apps.NewErrorCallResponse(errors.New("consent to use APIs and locations is required to install"))
			}

			_, out, err := a.proxy.InstallApp(
				proxy.NewIncomingFromContext(creq.Context),
				creq.Context, appID, true, secret)
			if err != nil {
				return apps.NewErrorCallResponse(errors.Wrap(err, "failed to install App"))
			}

			return mdResponse(out)
		},
	}
}

func (a *builtinApp) newInstallConsentForm(m apps.Manifest, creq apps.CallRequest) *apps.Form {
	fields := []apps.Field{}

	// Consent
	consent := ""
	if len(m.RequestedLocations) > 0 {
		consent += "\n- Add the following elements to the **Mattermost User Interface**:\n"
		for _, l := range m.RequestedLocations {
			consent += fmt.Sprintf("  - %s\n", l.Markdown())
		}
	}
	if len(m.RequestedPermissions) > 0 {
		consent += "- Access **Mattermost API** with the following permissions:\n"
		for _, permission := range m.RequestedPermissions {
			consent += fmt.Sprintf("  - %s\n", permission.String())
		}
	}
	if consent != "" {
		header := fmt.Sprintf("Application **%s** requires system administrator's consent to:\n\n", m.DisplayName)
		consent = header + consent + "---\n"

		fields = append(fields, apps.Field{
			Name:        fConsent,
			Type:        apps.FieldTypeBool,
			ModalLabel:  "Agree to grant the App access to APIs and Locations",
			Description: "",
			IsRequired:  true,
		})
	}

	// JWT secret
	if m.AppType == apps.AppTypeHTTP {
		fields = append(fields, apps.Field{
			Name:        fSecret,
			Type:        apps.FieldTypeText,
			ModalLabel:  "Outgoing JWT Secret",
			Description: fmt.Sprintf("The secret will be used to issue JWTs in outgoing messages to the app. Usually, it should be obtained from the App's web site, %s", m.HomepageURL),
			IsRequired:  false,
		})
	}

	var iconURL string
	if m.Icon != "" {
		iconURL = a.conf.Get().StaticURL(m.AppID, m.Icon)
	}

	return &apps.Form{
		Title:  fmt.Sprintf("Install App %s", m.DisplayName),
		Header: consent,
		Fields: fields,
		Call: &apps.Call{
			Path: pInstallConsent,
			Expand: &apps.Expand{
				AdminAccessToken: apps.ExpandAll,
			},
			State: m.AppID,
		},
		Icon: iconURL,
	}
}
