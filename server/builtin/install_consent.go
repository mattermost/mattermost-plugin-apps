// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

func (a *builtinApp) installConsentForm(creq *apps.CallRequest) *apps.CallResponse {
	appID, ok := creq.State.(apps.AppID)
	if !ok {
		return apps.NewErrorCallResponse(errors.New("no AppID found in State, don't know what to install"))
	}

	m, err := a.store.Manifest.Get(appID)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return formResponse(
		a.newInstallConsentForm(m, creq))
}

func (a *builtinApp) installConsentSubmit(creq *apps.CallRequest) *apps.CallResponse {
	deployType := apps.DeployType(creq.GetValue(fDeployType, ""))
	secret := creq.GetValue(fSecret, "")
	requireUserConsent := creq.BoolValue(fRequireUserConsent)
	locationsConsent := creq.BoolValue(fConsentLocations)
	permissionsConsent := creq.BoolValue(fConsentPermissions)
	appID, ok := creq.State.(apps.AppID)
	if !ok {
		return apps.NewErrorCallResponse(
			errors.New("no app ID in state, don't know what to install"))
	}

	m, err := a.store.Manifest.Get(appID)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "failed to load App manifest"))
	}

	if !locationsConsent && len(m.RequestedLocations) > 0 {
		return apps.NewErrorCallResponse(errors.New("consent to grant access to locations is required to install"))
	}
	if !permissionsConsent && len(m.RequestedPermissions) > 0 {
		return apps.NewErrorCallResponse(errors.New("consent to grant permissions is required to install"))
	}

	asAdmin := mmclient.AsAdmin(creq.Context)

	app, out, err := a.proxy.InstallApp(appID, asAdmin, "", creq.Context, !requireUserConsent, secret, deployType)
}

func (a *builtinApp) installConsentLookup(creq *apps.CallRequest) *apps.CallResponse {
	return nil
}

func (a *builtinApp) newConsentDeployTypeField(m *apps.Manifest, creq *apps.CallRequest) (field *apps.Field, selected apps.SelectOption) {
	opts := []apps.SelectOption{}
	for _, deployType := range m.DeployTypes() {
		_, canUse := a.proxy.CanDeploy(deployType)
		if canUse {
			opts = append(opts, apps.SelectOption{
				Label: deployType.String(),
				Value: string(deployType),
			})
		}
	}

	dtype := apps.DeployType(creq.GetValue(fDeployType, ""))
	defaultValue := apps.SelectOption{
		Label: dtype.String(),
		Value: string(dtype),
	}
	if len(opts) == 1 {
		defaultValue = opts[0]
	}

	return &apps.Field{
		Name:                fDeployType,
		Type:                apps.FieldTypeStaticSelect,
		IsRequired:          true,
		Description:         "Select how the App will be accessed.",
		Label:               "deploy-type",
		ModalLabel:          "Deployment method",
		SelectRefresh:       true,
		SelectStaticOptions: opts,
		Value:               defaultValue,
	}, defaultValue
}

func (a *builtinApp) newInstallConsentForm(m *apps.Manifest, creq *apps.CallRequest) *apps.Form {
	deployTypeField, selected := a.newConsentDeployTypeField(m, creq)
	fields := []*apps.Field{
		deployTypeField,
	}

	if deployType := apps.DeployType(selected.Value); deployType != "" {
		if len(m.RequestedLocations) > 0 {
			fields = append(fields, &apps.Field{
				Name:        fConsentLocations,
				Type:        apps.FieldTypeBool,
				ModalLabel:  fmt.Sprintf("Grant access to: %s", m.RequestedLocations),
				Description: "Check to allow the App to add user interface items in these locations",
				IsRequired:  true,
			})
		}
		if len(m.RequestedPermissions) > 0 {
			fields = append(fields, &apps.Field{
				Name:        fConsentPermissions,
				Type:        apps.FieldTypeBool,
				ModalLabel:  fmt.Sprintf("Grant permissions to: %s", m.RequestedPermissions),
				Description: "Check to allow the App to use these permissions",
				IsRequired:  true,
			})
		}
		fields = append(fields, &apps.Field{
			Name:        fRequireUserConsent,
			Type:        apps.FieldTypeBool,
			Label:       fmt.Sprintf("Require explicit user's consent to allow acting on behalf"),
			Description: "If off, users will be quietly connected to the App as needed; otherwise prompt for consent.",
		})

		if deployType == apps.DeployHTTP {
			fields = append(fields, &apps.Field{
				Name:        fSecret,
				Type:        apps.FieldTypeText,
				ModalLabel:  "Outgoing JWT Secret",
				Description: fmt.Sprintf("The secret will be used to issue JWTs in outgoing messages to the app. Usually, it should be obtained from the app's web site, %s", m.HomepageURL),
				IsRequired:  false,
			})
		}
	}

	return &apps.Form{
		Title:  fmt.Sprintf("Install App %s", m.DisplayName),
		Fields: fields,
		Call: &apps.Call{
			Path: pInstallConsent,
			Expand: &apps.Expand{
				AdminAccessToken: apps.ExpandAll,
			},
			State: m.AppID,
		},
	}
}
