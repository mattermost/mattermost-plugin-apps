package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	AppID          = apps.AppID("apps")
	AppDisplayName = "Mattermost Apps plugin"
	AppDescription = "Install and manage Mattermost Apps"
)

const (
	FieldAppID     = "app"
	FieldNamespace = "namespace"

	fAction         = "action"
	fAllowHTTPApps  = "allow_http_apps"
	fBase64         = "base64"
	fBase64Key      = "base64_key"
	fChannel        = "channel"
	fConsent        = "consent"
	fCount          = "count"
	fCreate         = "create"
	fCurrentValue   = "current_value"
	fDeployType     = "deploy_type"
	fDeveloperMode  = "developer_mode"
	fForce          = "force"
	fHashkeys       = "hashkeys"
	fID             = "id"
	fIncludePlugins = "include_plugins"
	fJSON           = "json"
	fLevel          = "level"
	fNewValue       = "new_value"
	fPage           = "page"
	fSecret         = "secret"
	fSessionID      = "session_id"
	fURL            = "url"
)

const (
	PathDebugClean        = "/debug/clean"
	PathDebugKVInfo       = "/debug/kv/info"
	PathDebugKVList       = "/debug/kv/list"
	PathDebugStoreList    = "/debug/store/list"
	PathDebugStorePollute = "/debug/store/pollute"
	PathDebugSessionsList = "/debug/session/list"
	pDebugBindings        = "/debug/bindings"
	pDebugKVClean         = "/debug/kv/clean"
	pDebugKVCreate        = "/debug/kv/create"
	pDebugKVEdit          = "/debug/kv/edit"
	pDebugKVEditModal     = "/debug/kv/edit-modal"
	pDebugLogs            = "/debug/logs"
	pDebugOAuthConfigView = "/debug/oauth/config/view"
	pDebugSessionsRevoke  = "/debug/session/delete"
	pDebugSessionsView    = "/debug/session/view"
	pDisable              = "/disable"
	pEnable               = "/enable"
	pInfo                 = "/info"
	pInstallConsent       = "/install-consent"
	pInstallConsentSource = "/install-consent/form"
	pInstallHTTP          = "/install-http"
	pInstallListed        = "/install-listed"
	pList                 = "/list"
	pUninstall            = "/uninstall"
	pSettings             = "/settings"
	pSettingsSave         = "/settings/save"
)

const (
	pLookupAppID     = "/q/app_id"
	pLookupNamespace = "/q/namespace"
)

type handler func(*incoming.Request, apps.CallRequest) apps.CallResponse

type builtinApp struct {
	conf           config.Service
	proxy          proxy.Service
	appservices    appservices.Service
	httpOut        httpout.Service
	sessionService session.Service
	router         map[string]handler
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(conf config.Service, proxy proxy.Service, appservices appservices.Service, httpOut httpout.Service, sessionService session.Service) *builtinApp {
	a := &builtinApp{
		conf:           conf,
		proxy:          proxy,
		appservices:    appservices,
		httpOut:        httpOut,
		sessionService: sessionService,
	}

	a.router = map[string]handler{
		// App's own bindings.
		appspath.Bindings: a.bindings,

		// Commands available to all users.
		pInfo: a.info,

		// Commands that require sysadmin.
		pDebugLogs:            requireAdmin(a.debugLogs),
		pDebugBindings:        requireAdmin(a.debugBindings),
		PathDebugClean:        requireAdmin(a.debugClean),
		pDebugKVClean:         requireAdmin(a.debugKVClean),
		pDebugKVCreate:        requireAdmin(a.debugKVCreate),
		pDebugKVEdit:          requireAdmin(a.debugKVEdit),
		PathDebugKVInfo:       requireAdmin(a.debugKVInfo),
		PathDebugKVList:       requireAdmin(a.debugKVList),
		PathDebugStoreList:    requireAdmin(a.debugStoreList),
		PathDebugStorePollute: requireAdmin(a.debugStorePollute),
		PathDebugSessionsList: requireAdmin(a.debugSessionsList),
		pDebugSessionsRevoke:  requireAdmin(a.debugSessionsRevoke),
		pDebugSessionsView:    requireAdmin(a.debugSessionsView),
		pDebugOAuthConfigView: requireAdmin(a.debugOAuthConfigView),
		pEnable:               requireAdmin(a.enable),
		pDisable:              requireAdmin(a.disable),
		pInstallListed:        requireAdmin(a.installListed),
		pInstallHTTP:          requireAdmin(a.installHTTP),
		pList:                 requireAdmin(a.list),
		pUninstall:            requireAdmin(a.uninstall),
		pSettings:             requireAdmin(a.settings),

		// Modals.
		pDebugKVEditModal:     requireAdmin(a.debugKVEditModal),
		pInstallConsent:       requireAdmin(a.installConsent),
		pInstallConsentSource: requireAdmin(a.installConsentForm),
		pSettingsSave:         requireAdmin(a.settingsSave),

		// Lookups.
		pLookupAppID:     requireAdmin(a.lookupAppID),
		pLookupNamespace: requireAdmin(a.lookupNamespace),
	}

	return a
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:                AppID,
		Version:              apps.AppVersion(conf.PluginManifest.Version),
		DisplayName:          AppDisplayName,
		Description:          AppDescription,
		Deploy:               apps.Deploy{},
		RequestedPermissions: apps.Permissions{
			// apps.PermissionActAsBot,
			// apps.PermissionActAsUser,
		},
		RequestedLocations: apps.Locations{
			apps.LocationCommand,
		},
		Bindings: &apps.Call{
			Path: appspath.Bindings,
			Expand: &apps.Expand{
				ActingUser: apps.ExpandSummary,
				Locale:     apps.ExpandAll,
			},
		},
	}
}

func App(conf config.Config) apps.App {
	m := Manifest(conf)
	return apps.App{
		Manifest:           m,
		DeployType:         apps.DeployBuiltin,
		BotUserID:          conf.BotUserID,
		BotUsername:        config.BotUsername,
		GrantedLocations:   m.RequestedLocations,
		GrantedPermissions: m.RequestedPermissions,
	}
}

func (a *builtinApp) Roundtrip(ctx context.Context, _ apps.App, creq apps.CallRequest, async bool) (out io.ReadCloser, err error) {
	self := App(a.conf.Get())
	r := a.proxy.NewIncomingRequest().WithCtx(ctx).WithDestination(self.AppID)

	if creq.Context.ActingUser != nil {
		r = r.WithActingUserID(creq.Context.ActingUser.Id)
	}

	defer func(log utils.Logger) {
		if x := recover(); x != nil {
			stack := string(debug.Stack())
			txt := "Call `" + creq.Path + "` panic-ed."
			log = log.With(
				creq,
				"error", x,
				"stack", stack,
			)
			if creq.RawCommand != "" {
				txt = "Command `" + creq.RawCommand + "` panic-ed."
				log.Errorw("Recovered from a panic in a command", "command", creq.RawCommand)
			} else {
				log.Errorf("Recovered from a panic in a Call")
			}

			if a.conf.Get().DeveloperMode {
				txt += "\n"
				txt += fmt.Sprintf("Error: **%v**.\n", x)
				txt += fmt.Sprintf("Stack:\n%v", utils.CodeBlock(stack))
			} else {
				txt += "Please check the server logs for more details."
			}
			out = nil
			data, errr := json.Marshal(apps.NewTextResponse(txt))
			if errr != nil {
				err = errr
				return
			}
			err = nil
			out = io.NopCloser(bytes.NewReader(data))
		}
	}(r.Log)

	readcloser := func(cresp apps.CallResponse) (io.ReadCloser, error) {
		data, err := json.Marshal(cresp)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	loc := a.newLocalizer(creq)
	confErr := a.checkConfigValid(loc)
	if confErr != nil {
		return readcloser(apps.NewErrorResponse(confErr))
	}

	h, ok := a.router[creq.Path]
	if !ok {
		return nil, utils.NewNotFoundError(creq.Path)
	}
	return readcloser(h(r, creq))
}

func (a *builtinApp) GetStatic(_ context.Context, _ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.NewNotFoundError("static support is not implemented")
}

func requireAdmin(h handler) handler {
	return func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
		if creq.Context.ActingUser == nil {
			return apps.NewErrorResponse(utils.NewInvalidError(
				"no or invalid ActingUser in the context for %s, please make sure Expand.ActingUser is set", creq.Path))
		}
		if !creq.Context.ActingUser.IsSystemAdmin() {
			return apps.NewErrorResponse(utils.NewUnauthorizedError(
				"user %s (%s) is not a sysadmin", creq.Context.ActingUser.GetDisplayName(model.ShowUsername), creq.Context.ActingUser.Id))
		}
		return h(r, creq)
	}
}

func (a *builtinApp) newLocalizer(creq apps.CallRequest) *i18n.Localizer {
	if creq.Context.Locale != "" {
		return i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	}

	return a.conf.I18N().GetUserLocalizer(creq.Context.ActingUser.Id)
}

func (a *builtinApp) checkConfigValid(loc *i18n.Localizer) error {
	oauthEnabled := a.conf.MattermostConfig().Config().ServiceSettings.EnableOAuthServiceProvider

	if oauthEnabled == nil || !*oauthEnabled {
		integrationManagementPage := fmt.Sprintf("%s/admin_console/integrations/integration_management", a.conf.Get().MattermostSiteURL)

		message := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "command.error.oauth2.disabled",
				Other: "The system setting `Enable OAuth 2.0 Service Provider` needs to be enabled in order for the Apps plugin to work. Please go to {{.IntegrationManagementPage}} and enable it.",
			},
			TemplateData: map[string]string{
				"IntegrationManagementPage": integrationManagementPage,
			},
		})

		return errors.New(message)
	}

	return nil
}
