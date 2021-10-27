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
			m, err := a.stateAsManifest(creq)
			if err != nil {
				return nil, errors.Wrap(err, "failed to find a valid manifest in State")
			}
			return a.newInstallConsentForm(*m, creq, ""), nil
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			deployType := apps.DeployType(creq.GetValue(fDeployType, ""))
			secret := creq.GetValue(fSecret, "")
			consent := creq.BoolValue(fConsent)
			m, err := a.stateAsManifest(creq)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to find a valid manifest in State"))
			}
			if !consent && len(m.RequestedLocations)+len(m.RequestedPermissions) > 0 {
				return apps.NewErrorResponse(errors.New("consent to use APIs and locations is required to install"))
			}

			_, out, err := a.proxy.InstallApp(
				proxy.NewIncomingFromContext(creq.Context),
				creq.Context, m.AppID, deployType, true, secret)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to install App"))
			}

			return apps.NewTextResponse(out)
		},
	}
}

func (a *builtinApp) newConsentDeployTypeField(m apps.Manifest, creq apps.CallRequest, requestedType apps.DeployType) (
	apps.Field, apps.DeployType) {
	opts := []apps.SelectOption{}

	// See if there's user selection for the current value of the field.
	requestedType = apps.DeployType(creq.GetValue(fDeployType, string(requestedType)))

	var selectedValue *apps.SelectOption
	var selectedType apps.DeployType
	for _, typ := range m.DeployTypes() {
		_, canUse := a.proxy.CanDeploy(typ)
		if canUse {
			opt := apps.SelectOption{
				Label: typ.String(),
				Value: string(typ),
			}
			opts = append(opts, opt)
			if typ == requestedType {
				selectedValue = &opt
				selectedType = typ
			}
		}
	}
	if len(opts) == 1 {
		selectedValue = &opts[0]
	}

	return apps.Field{
		Name:                fDeployType,
		Type:                apps.FieldTypeStaticSelect,
		IsRequired:          true,
		Description:         "Select how the App will be accessed.",
		Label:               "deploy-type",
		ModalLabel:          "Deployment method",
		SelectRefresh:       true,
		SelectStaticOptions: opts,
		Value:               selectedValue,
	}, selectedType
}

func (a *builtinApp) newInstallConsentForm(m apps.Manifest, creq apps.CallRequest, deployType apps.DeployType) *apps.Form {
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

		value := creq.BoolValue(fConsent)
		fields = append(fields, apps.Field{
			Name:        fConsent,
			Type:        apps.FieldTypeBool,
			ModalLabel:  "Agree to grant the App access to APIs and Locations",
			Description: "",
			IsRequired:  true,
			Value:       value,
		})
	}

	deployTypeField, deployType := a.newConsentDeployTypeField(m, creq, deployType)
	fields = append(fields, deployTypeField)

	// JWT secret
	if deployType == apps.DeployHTTP && m.Contains(apps.DeployHTTP) && m.HTTP.UseJWT {
		fields = append(fields, apps.Field{
			Name:        fSecret,
			Type:        apps.FieldTypeText,
			ModalLabel:  "Outgoing JWT Secret",
			Description: fmt.Sprintf("The secret will be used to issue JWTs in outgoing messages to the app. Usually, it should be obtained from the App's web site, %s", m.HomepageURL),
			IsRequired:  false,
		})
	}

	// TODO: figure out a way to access the static assets before the app is installed
	// var iconURL string
	// if m.Icon != "" {
	// 	iconURL = a.conf.Get().StaticURL(m.AppID, m.Icon)
	// }

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
		// Icon: iconURL, see above TODO
	}
}

func (a *builtinApp) stateAsManifest(creq apps.CallRequest) (*apps.Manifest, error) {
	id, ok := creq.State.(string)
	if !ok {
		return nil, errors.New("no app ID in State, don't know what to install")
	}
	appID := apps.AppID(id)

	return a.proxy.GetManifest(appID)
}
