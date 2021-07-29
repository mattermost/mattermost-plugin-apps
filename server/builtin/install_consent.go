// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) installConsentForm(creq *apps.CallRequest) *apps.CallResponse {
	// s, ok := c.Context.Props[contextInstallAppID]
	// if !ok {
	// 	return apps.NewErrorCallResponse(errors.New("no AppID to install in Context"))
	// }
	// appID := apps.AppID(s)

	// m, err := a.store.Manifest.Get(appID)
	// if err != nil {
	// 	return apps.NewErrorCallResponse(err)
	// }

	// return a.installAppFormManifest(m, c)
	return nil
}

func (a *builtinApp) installConsentSubmit(creq *apps.CallRequest) *apps.CallResponse {
	// secret := call.GetStringValue(fSecret, "")
	// requireUserConsent := call.GetBoolValue(fRequireUserConsent)
	// locationsConsent := call.GetBoolValue(fConsentLocations)
	// permissionsConsent := call.GetBoolValue(fConsentPermissions)
	// id := ""
	// if v, _ := call.Context.Props[contextInstallAppID]; v != "" {
	// 	id = v
	// }

	// m, err := a.store.Manifest.Get(apps.AppID(id))
	// if err != nil {
	// 	return apps.NewErrorCallResponse(errors.Wrap(err, "failed to load App manifest"))
	// }

	// if !locationsConsent && len(m.RequestedLocations) > 0 {
	// 	return apps.NewErrorCallResponse(errors.New("consent to grant access to UI locations is required to install"))
	// }
	// if !permissionsConsent && len(m.RequestedPermissions) > 0 {
	// 	return apps.NewErrorCallResponse(errors.New("consent to grant permissions is required to install"))
	// }

	// app, out, err := a.proxy.InstallApp(call.Context, &apps.InInstallApp{
	// 	Manifest:         m,
	// 	OAuth2TrustedApp: !requireUserConsent,
	// 	Secret:           secret,
	// })
	// return apps.NewCallResponse(out, app, err)
	return nil
}

func (a *builtinApp) installConsentLookup(creq *apps.CallRequest) *apps.CallResponse {
	return nil
}

// func (a *builtinApp) generateConsentAppTypeField(m *apps.Manifest, creq *apps.CallRequest) (field *apps.Field, selected bool) {
// 	v, ok := creq.Values[fAppType]

// 	field := &apps.Field{
// 		Name: fAppType,
// 		Type: apps.FieldTypeStaticSelect,
// 		IsRequired: true,
// 		Description : "Select where the App resides (App type).",
// 		Label: "app-type",
// 		ModalLabel: "Access method",
// 		SelectRefresh: true,
// 	}

// 	if app.Manifest
// 		// SelectStaticOptions is the list of options to display in a static select
// 		// field.
// 		SelectStaticOptions []SelectOption `json:"options,omitempty"`

// 		// Text props
// 		TextSubtype   TextFieldSubtype `json:"subtype,omitempty"`
// 		TextMinLength int              `json:"min_length,omitempty"`
// 		TextMaxLength int              `json:"max_length,omitempty"`

// 	}
func (a *builtinApp) newInstallConsentForm(m *apps.Manifest, creq *apps.CallRequest) *apps.Form {

	// 	fields := []*apps.Field{}

	// 	if len(m.RequestedLocations) > 0 {
	// 		fields = append(fields, &apps.Field{
	// 			Name:        fConsentLocations,
	// 			Type:        apps.FieldTypeBool,
	// 			Label:       "Application may display its UI elements in the following locations",
	// 			Description: fmt.Sprintf("%s", m.RequestedLocations),
	// 			IsRequired:  true,
	// 		})
	// 	}

	// 	if len(m.RequestedPermissions) > 0 {
	// 		fields = append(fields, &apps.Field{
	// 			Name:        fConsentPermissions,
	// 			Type:        apps.FieldTypeBool,
	// 			Label:       "Application will have the following permissions",
	// 			Description: fmt.Sprintf("%s", m.RequestedPermissions),
	// 			IsRequired:  true,
	// 		})
	// 	}

	// 	fields = append(fields, &apps.Field{
	// 		Name:        fRequireUserConsent,
	// 		Type:        apps.FieldTypeBool,
	// 		Label:       fmt.Sprintf("Require explicit user's consent to allow %s App impersonate the user", m.AppID),
	// 		Description: "If off, users will be quietly connected to the App as needed; otherwise prompt for consent.",
	// 	})

	// 	if m.Type == apps.AppTypeHTTP {
	// 		fields = append(fields, &apps.Field{
	// 			Name:             fSecret,
	// 			Type:             apps.FieldTypeText,
	// 			Description:      "The App's secret to use in JWT.",
	// 			Label:            fSecret,
	// 			AutocompleteHint: "paste the secret obtained from the App",
	// 			IsRequired:       true,
	// 		})
	// 	}

	// 	cr := &apps.CallResponse{
	// 		Type: apps.CallResponseTypeForm,
	// 		Form: &apps.Form{
	// 			Title:  fmt.Sprintf("Install App %s", m.DisplayName),
	// 			Fields: fields,
	// 			Call: &apps.Call{
	// 				Path: pInstallConsent,
	// 				Expand: &apps.Expand{
	// 					AdminAccessToken: apps.ExpandAll,
	// 				},
	// 				State: map[string]string{
	// 					contextInstallAppID: string(m.AppID),
	// 				},
	// 			},
	// 		},
	// 	}
	// 	return cr
	return nil
}
