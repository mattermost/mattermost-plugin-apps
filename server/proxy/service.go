// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

type Proxy struct {
	callOnceMutex *cluster.Mutex

	builtinUpstreams map[apps.AppID]upstream.Upstream

	mm        *pluginapi.Client
	log       utils.Logger
	conf      config.Service
	store     *store.Service
	httpOut   httpout.Service
	upstreams sync.Map // key: apps.AppID, value upstream.Upstream
}

type Service interface {
	// To update on configuration changes
	config.Configurable

	// Admin REST API methods.
	DisableApp(Incoming, apps.AppID) (md.MD, error)
	EnableApp(Incoming, apps.AppID) (md.MD, error)
	InstallApp(_ Incoming, _ apps.AppID, _ apps.DeployType, trusted bool, secret string) (*apps.App, md.MD, error)
	UninstallApp(Incoming, apps.AppID) (md.MD, error)

	// REST API methods used by user agents (mobile, desktop, web).
	Call(Incoming, apps.AppID, apps.CallRequest) apps.ProxyCallResponse
	CompleteRemoteOAuth2(_ Incoming, _ apps.AppID, urlValues map[string]interface{}) error
	GetBindings(Incoming, apps.Context) ([]apps.Binding, error)
	GetRemoteOAuth2ConnectURL(Incoming, apps.AppID) (string, error)
	GetStatic(_ apps.AppID, path string) (io.ReadCloser, int, error)

	// User-less notification sinks.
	Notify(cc apps.Context, subj apps.Subject) error
	NotifyRemoteWebhook(app apps.App, data []byte, path string) error

	// Internal go API used by other packages.
	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
	AddLocalManifest(m apps.Manifest) (md.MD, error)
	CanDeploy(deployType apps.DeployType) (allowed, usable bool)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []apps.App
	GetListedApps(filter string, includePluginApps bool) []apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	GetManifestFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error)
	SynchronizeInstalledApps() error
}

var _ Service = (*Proxy)(nil)

func NewService(mm *pluginapi.Client, log utils.Logger, conf config.Service, store *store.Service, mutex *cluster.Mutex, httpOut httpout.Service) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]upstream.Upstream{},
		mm:               mm,
		log:              log,
		conf:             conf,
		store:            store,
		callOnceMutex:    mutex,
		httpOut:          httpOut,
	}
}

func (p *Proxy) Configure(conf config.Config) error {
	newUpstream := func(dtype apps.DeployType, makeUpstreamf func() (upstream.Upstream, error)) {
		allowed, _ := p.CanDeploy(dtype)
		if !allowed {
			p.upstreams.Delete(dtype)
			return
		}
		// Override whatever might have been stored before.
		up, err := makeUpstreamf()
		if err != nil {
			p.upstreams.Delete(dtype)
			p.log.WithError(err).Debugw("failed to initialize upstream", "deploy_type", dtype)
		}
		p.upstreams.Store(dtype, up)
	}

	newUpstream(apps.DeployHTTP, func() (upstream.Upstream, error) {
		return uphttp.NewUpstream(p.httpOut), nil
	})
	newUpstream(apps.DeployAWSLambda, func() (upstream.Upstream, error) {
		return upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, p.log)
	})
	newUpstream(apps.DeployPlugin, func() (upstream.Upstream, error) {
		return upplugin.NewUpstream(&p.mm.Plugin), nil
	})
	newUpstream(apps.DeployKubeless, func() (upstream.Upstream, error) {
		return upkubeless.MakeUpstream()
	})
	return nil
}

// CanDeploy returns the availability of deployType.  allowed indicates that the
// type can be used in the current configuration. usable indicates that it is
// configured and can be accessed, or deployed to.
func (p *Proxy) CanDeploy(deployType apps.DeployType) (allowed, usable bool) {
	_, usable = p.upstreams.Load(deployType)

	conf := p.conf.GetConfig()
	supportedTypes := []apps.DeployType{}

	// Initialize with the set supported in all configurations.
	supportedTypes = append(supportedTypes,
		apps.DeployAWSLambda,
		apps.DeployBuiltin,
		apps.DeployPlugin,
	)

	switch {
	case conf.DeveloperMode:
		// In dev mode support any deploy type.
		return true, usable

	case conf.MattermostCloudMode:
		// Nothing else in Mattermost Cloud mode.

	case !conf.MattermostCloudMode:
		// Add more deploy types in self-managed mode.
		supportedTypes = append(supportedTypes,
			apps.DeployHTTP,
			apps.DeployKubeless)

	default:
		return false, false
	}

	for _, t := range supportedTypes {
		if deployType == t {
			return true, usable
		}
	}
	return false, false
}

func CanDeploy(p Service, deployType apps.DeployType) error {
	_, canDeploy := p.CanDeploy(deployType)
	if !canDeploy {
		return errors.Errorf("%s app deployment is not configured on this instance of Mattermost", deployType)
	}
	return nil
}

func (p *Proxy) AddBuiltinUpstream(appID apps.AppID, up upstream.Upstream) {
	if p.builtinUpstreams == nil {
		p.builtinUpstreams = map[apps.AppID]upstream.Upstream{}
	}
	p.builtinUpstreams[appID] = up
	p.store.App.InitBuiltin()
}

func WriteCallError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(apps.CallResponse{
		Type:      apps.CallResponseTypeError,
		ErrorText: err.Error(),
	})
}
