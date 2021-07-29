package builtin

import (
	"bytes"
	"encoding/json"
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
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
	"github.com/pkg/errors"
)

const (
	AppID          = "apps"
	AppDisplayName = "Mattermost Apps plugin"
	AppDescription = "Install and manage Mattermost Apps"
)

const (
	contextInstallAppID = "install_app_id"

	fURL                = "url"
	fConsentPermissions = "consent_permissions"
	fConsentLocations   = "consent_locations"
	fRequireUserConsent = "require_user_consent"
	fSecret             = "secret"
	fAppID              = "app"
	fIncludePlugins     = "include_plugins"
	fDeployType         = "deploy_type"
	fUserID             = "user"
)

const (
	pDebugBindings  = "/debug-bindings"
	pDebugClean     = "/debug-clean"
	pInfo           = "/info"
	pList           = "/list"
	pInstallURL     = "/install-url"
	pInstallS3      = "/install-s3"
	pInstallConsent = "/install-consent"
)

type builtinApp struct {
	conf    config.Service
	mm      *pluginapi.Client
	log     utils.Logger
	proxy   proxy.Service
	store   *store.Service
	httpOut httpout.Service
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(mm *pluginapi.Client, log utils.Logger, conf config.Service, proxy proxy.Service, store *store.Service, httpOut httpout.Service) *builtinApp {
	return &builtinApp{
		mm:      mm,
		log:     log,
		conf:    conf,
		proxy:   proxy,
		store:   store,
		httpOut: httpOut,
	}
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:       AppID,
		Version:     apps.AppVersion(conf.BuildConfig.BuildHashShort),
		DisplayName: AppDisplayName,
		Description: AppDescription,
	}
}

func App(conf config.Config) *apps.App {
	return &apps.App{
		Manifest:    Manifest(conf),
		DeployType:  apps.DeployBuiltin,
		BotUserID:   conf.BotUserID,
		BotUsername: config.BotUsername,
		GrantedLocations: apps.Locations{
			apps.LocationCommand,
		},
	}
}

func (a *builtinApp) Roundtrip(_ *apps.App, creq *apps.CallRequest, async bool) (io.ReadCloser, error) {
	var f func(*apps.CallRequest) *apps.CallResponse

	switch creq.Path {
	case apps.DefaultBindings.Path:
		f = a.getBindings

	case formPath(pInfo),
		formPath(pDebugClean):
		f = emptyForm

	case submitPath(pInfo):
		f = a.info
	case submitPath(pDebugClean):
		f = a.debugClean

	case formPath(pList):
		f = a.listForm
	case submitPath(pList):
		f = a.list

	case formPath(pInstallS3):
		f = a.installS3Form
	case submitPath(pInstallS3):
		f = a.installS3Submit
	case lookupPath(pInstallS3):
		f = a.installLookup

	case formPath(pInstallURL):
		f = a.installURLForm
	case submitPath(pInstallURL):
		f = a.installURLSubmit

	case formPath(pInstallConsent):
		f = a.installConsentForm
	case lookupPath(pInstallConsent):
		f = a.installConsentLookup
	case submitPath(pInstallConsent):
		f = a.installConsentSubmit

	default:
		return nil, errors.Errorf("%s is not found", creq.Path)
	}

	cresp := f(creq)
	data, err := json.Marshal(cresp)
	if err != nil {
		return nil, err
	}
	a.log.Debugf("<>/<> RT: " + string(data))
	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func (a *builtinApp) GetStatic(_ *apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.ErrNotFound
}

func mdResponse(format string, args ...interface{}) *apps.CallResponse {
	return &apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: md.Markdownf(format, args...),
	}
}

func formResponse(form *apps.Form) *apps.CallResponse {
	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: form,
	}
}

func dataResponse(data interface{}) *apps.CallResponse {
	return &apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: data,
	}
}

func emptyForm(_ *apps.CallRequest) *apps.CallResponse {
	return formResponse(&apps.Form{})
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
