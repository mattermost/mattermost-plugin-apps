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

func (a *builtinApp) installConsentForm(creq apps.CallRequest) apps.CallResponse {
	loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	m, err := a.stateAsManifest(creq)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to find a valid manifest in State"))
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq, "", loc))
}

func (a *builtinApp) installConsent(creq apps.CallRequest) apps.CallResponse {
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
		Name:       fDeployType,
		Type:       apps.FieldTypeStaticSelect,
		IsRequired: true,
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.deploy_type.description",
			Other: "Select how the App will be accessed.",
		}),
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.deploy_type.label",
			Other: "deploy-type",
		}),
		ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.deploy_type.modal_label",
			Other: "Deployment method",
		}),
		SelectRefresh:       true,
		SelectOptions: opts,
		Value:               selectedValue,
	}, selectedType
}

func (a *builtinApp) newInstallConsentForm(m apps.Manifest, creq apps.CallRequest, deployType apps.DeployType, loc *i18n.Localizer) apps.Form {
	fields := []apps.Field{}

	// Consent
	h := ""
	if len(m.RequestedLocations) > 0 {
		h += a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.install_consent.header.locations",
			Other: "\n- Add the following elements to the **Mattermost User Interface**:\n",
		})
		// (Mattermost) locations themselves are not localized
		for _, l := range m.RequestedLocations {
			h += fmt.Sprintf("  - %s\n", l.Markdown())
		}
	}
	if len(m.RequestedPermissions) > 0 {
		h += a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.install_consent.header.permissions",
			Other: "- Access **Mattermost API** with the following permissions:\n",
		})
		// Permissions are not localized
		for _, permission := range m.RequestedPermissions {
			h += fmt.Sprintf("  - %s\n", permission.String())
		}
	}
	if h != "" {
		header := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "modal.install_consent.header.header",
				Other: "Application **{{.DisplayName}}** requires system administrator's consent to:\n\n",
			},
			TemplateData: map[string]string{
				"DisplayName": m.DisplayName,
			},
		})
		h = header + h + "---\n"

		value := creq.BoolValue(fConsent)
		fields = append(fields, apps.Field{
			Name: fConsent,
			Type: apps.FieldTypeBool,
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "field.consent.modal_label",
				Other: "Agree to grant the app access to APIs and Locations",
			}),
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
			Name: fSecret,
			Type: apps.FieldTypeText,
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "field.secret.modal_label.use_jwt",
				Other: "Outgoing JWT Secret",
			}),

			Description: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "field.secret.description.use_jwt",
					Other: "The secret will be used to issue JWTs in outgoing messages to the app. Usually, it should be obtained from the App's web site, {{.HomepageURL}}",
				},
				TemplateData: map[string]string{
					"HomepageURL": m.HomepageURL,
				},
			}),
			IsRequired: false,
		})
	}

	// TODO: figure out a way to access the static assets before the app is installed
	// var iconURL string
	// if m.Icon != "" {
	// 	iconURL = a.conf.Get().StaticURL(m.AppID, m.Icon)
	// }

	return apps.Form{
		Title: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "modal.install_consent.title",
				Other: "Install App {{.DisplayName}}",
			},
			TemplateData: map[string]string{
				"DisplayName": m.DisplayName,
			},
		}),
		Header: h,
		Fields: fields,
		Submit: newUserCall(pInstallConsent).WithState(m.AppID),
		Source: newUserCall(pInstallConsentForm).WithState(m.AppID),
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
