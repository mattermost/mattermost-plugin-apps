// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func (a *builtinApp) installConsent() handler {
	return handler{
		requireSysadmin: true,

		formf: func(creq apps.CallRequest) (*apps.Form, error) {
			loc := a.newLocalizer(creq)
			m, err := a.stateAsManifest(creq)
			if err != nil {
				return nil, errors.Wrap(err, "failed to find a valid manifest in State")
			}
			return a.newInstallConsentForm(*m, creq, "", loc), nil
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

func (a *builtinApp) newConsentDeployTypeField(m apps.Manifest, creq apps.CallRequest, requestedType apps.DeployType, loc *i18n.Localizer) (
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
		Description:         a.conf.Local(loc, "field.deploy_type.description"),
		Label:               a.conf.Local(loc, "field.deploy_type.label"),
		ModalLabel:          a.conf.Local(loc, "field.deploy_type.modal_label"),
		SelectRefresh:       true,
		SelectStaticOptions: opts,
		Value:               selectedValue,
	}, selectedType
}

func (a *builtinApp) newInstallConsentForm(m apps.Manifest, creq apps.CallRequest, deployType apps.DeployType, loc *i18n.Localizer) *apps.Form {
	fields := []apps.Field{}

	// Consent
	h := ""
	if len(m.RequestedLocations) > 0 {
		h += fmt.Sprintf("\n%s\n", a.conf.Local(loc, "modal.install_consent.header.locations"))
		// (Mattermost) locations themselves are not localized
		for _, l := range m.RequestedLocations {
			h += fmt.Sprintf("  - %s\n", l.Markdown())
		}
	}
	if len(m.RequestedPermissions) > 0 {
		h += fmt.Sprintf("%s\n", a.conf.Local(loc, "modal.install_consent.header.permissions"))
		// Permissions are not localized
		for _, permission := range m.RequestedPermissions {
			h += fmt.Sprintf("  - %s\n", permission.String())
		}
	}
	if h != "" {
		header := fmt.Sprintf("%s\n\n",
			a.conf.LocalWithTemplate(loc, "modal.install_consent.header.header",
				map[string]string{"DisplayName": m.DisplayName}))
		h = header + h + "---\n"

		value := creq.BoolValue(fConsent)
		fields = append(fields, apps.Field{
			Name:        fConsent,
			Type:        apps.FieldTypeBool,
			ModalLabel:  a.conf.Local(loc, "field.consent.modal_label"),
			Description: "",
			IsRequired:  true,
			Value:       value,
		})
	}

	deployTypeField, deployType := a.newConsentDeployTypeField(m, creq, deployType, loc)
	fields = append(fields, deployTypeField)

	// JWT secret
	if deployType == apps.DeployHTTP && m.Contains(apps.DeployHTTP) && m.HTTP.UseJWT {
		fields = append(fields, apps.Field{
			Name:       fSecret,
			Type:       apps.FieldTypeText,
			ModalLabel: a.conf.Local(loc, "field.secret.modal_label.use_jwt"),
			Description: a.conf.LocalWithTemplate(loc, "field.secret.description.use_jwt",
				map[string]string{"HomepageURL": m.HomepageURL}),
			IsRequired: false,
		})
	}

	// TODO: figure out a way to access the static assets before the app is installed
	// var iconURL string
	// if m.Icon != "" {
	// 	iconURL = a.conf.Get().StaticURL(m.AppID, m.Icon)
	// }

	return &apps.Form{
		Title: a.conf.LocalWithTemplate(loc, "modal.install_consent.title",
			map[string]string{"DisplayName": m.DisplayName}),
		Header: h,
		Fields: fields,
		Call: &apps.Call{
			Path: pInstallConsent,
			Expand: &apps.Expand{
				AdminAccessToken: apps.ExpandAll,
				ActingUser:       apps.ExpandSummary,
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
