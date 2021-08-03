package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	AppID          = "apps"
	AppDisplayName = "Mattermost Apps plugin"
	AppDescription = "Install and manage Mattermost Apps"
)

const (
	fURL            = "url"
	fConsent        = "consent"
	fSecret         = "secret"
	fAppID          = "app"
	fVersion        = "version"
	fIncludePlugins = "include_plugins"
	fDeployType     = "deploy_type"
	fUserID         = "user"
)

const (
	pDebugBindings      = "/debug-bindings"
	pDebugClean         = "/debug-clean"
	pInfo               = "/info"
	pList               = "/list"
	pUninstall          = "/uninstall"
	pEnable             = "/enable"
	pDisable            = "/disable"
	pInstallURL         = "/install-url"
	pInstallS3          = "/install-s3"
	pInstallMarketplace = "/install-marketplace"
	pInstallConsent     = "/install-consent"
)

type builtinApp struct {
	conf    config.Service
	mm      *pluginapi.Client
	log     utils.Logger
	proxy   proxy.Service
	store   *store.Service
	httpOut httpout.Service

	router map[string]func(apps.CallRequest) apps.CallResponse
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(mm *pluginapi.Client, log utils.Logger, conf config.Service, proxy proxy.Service, store *store.Service, httpOut httpout.Service) *builtinApp {
	a := &builtinApp{
		mm:      mm,
		log:     log,
		conf:    conf,
		proxy:   proxy,
		store:   store,
		httpOut: httpOut,
		router:  map[string]func(apps.CallRequest) apps.CallResponse{},
	}

	a.router[apps.DefaultBindings.Path] = a.getBindings

	a.route(pDebugBindings, a.debugBindings, nil, nil)
	a.route(pDebugClean, a.debugClean, nil, nil)
	a.route(pInfo, a.info, nil, nil)
	a.route(pList, a.list, a.listForm, nil)

	a.route(pDisable, a.disableSubmit, a.disableForm, a.disableLookup)
	a.route(pEnable, a.enableSubmit, a.enableForm, a.enableLookup)
	a.route(pInstallConsent, a.installConsentSubmit, a.installConsentForm, nil)
	a.route(pInstallMarketplace, a.installMarketplaceSubmit, a.installMarketplaceForm, a.installMarketplaceLookup)
	a.route(pInstallS3, a.installS3Submit, a.installS3Form, a.installS3Lookup)
	a.route(pInstallURL, a.installURLSubmit, a.installURLForm, nil)
	a.route(pUninstall, a.uninstallSubmit, a.uninstallForm, a.uninstallLookup)

	return a
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:       AppID,
		Version:     apps.AppVersion(conf.BuildConfig.BuildHashShort),
		DisplayName: AppDisplayName,
		Description: AppDescription,
	}
}

func App(conf config.Config) apps.App {
	return apps.App{
		Manifest:    Manifest(conf),
		DeployType:  apps.DeployBuiltin,
		BotUserID:   conf.BotUserID,
		BotUsername: config.BotUsername,
		GrantedLocations: apps.Locations{
			apps.LocationCommand,
		},
		GrantedPermissions: apps.Permissions{
			apps.PermissionActAsAdmin,
		},
	}
}

func (a *builtinApp) Roundtrip(_ apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
	f, ok := a.router[creq.Path]
	if !ok {
		return nil, utils.NewNotFoundError("%s is not found", creq.Path)
	}
	cresp := f(creq)
	data, err := json.Marshal(cresp)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func (a *builtinApp) GetStatic(_ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.ErrNotFound
}

func mdResponse(format string, args ...interface{}) apps.CallResponse {
	return apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: fmt.Sprintf(format, args...),
	}
}

func formResponse(form apps.Form) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &form,
	}
}

func dataResponse(data interface{}) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: data,
	}
}

func emptyForm(_ apps.CallRequest) apps.CallResponse {
	return formResponse(apps.Form{})
}

func submitPath(p string) string {
	return path.Join(p, "submit")
}

func formPath(p string) string {
	return path.Join(p, "form")
}

func lookupPath(p string) string {
	return path.Join(p, "lookup")
}

func (a *builtinApp) route(path string, submitf, formf, lookupf func(apps.CallRequest) apps.CallResponse) {
	a.router[submitPath(path)] = submitf

	if formf == nil {
		formf = emptyForm
	}
	a.router[formPath(path)] = formf

	if lookupf != nil {
		a.router[lookupPath(path)] = lookupf
	}
}
