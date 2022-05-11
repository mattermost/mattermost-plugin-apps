package appclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	HeaderEtagClient = "If-None-Match"
	HeaderAuth       = "Authorization"

	AppsPluginName = "com.mattermost.apps"
)

type ClientPP struct {
	URL        string       // The location of the server, for example  "http://localhost:8065"
	HTTPClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	HTTPHeader map[string]string // Headers to be copied over for each request

	// TrueString is the string value sent to the server for true boolean query parameters.
	trueString string

	// FalseString is the string value sent to the server for false boolean query parameters.
	falseString string

	fromPlugin bool
}

func NewAppsPluginAPIClient(url string) *ClientPP {
	url = strings.TrimRight(url, "/")
	return &ClientPP{url, &http.Client{}, "", "", map[string]string{}, "", "", false}
}

func NewAppsPluginAPIClientFromPluginAPI(api upplugin.PluginHTTPAPI) *ClientPP {
	httpClient := upplugin.MakePluginHTTPClient(api)

	return &ClientPP{"", &httpClient, "", "", map[string]string{}, "", "", true}
}

func (c *ClientPP) SetOAuthToken(token string) {
	c.AuthToken = token
	c.AuthType = model.HeaderBearer
}

func (c *ClientPP) KVSet(prefix, id string, in interface{}) (bool, *model.Response, error) {
	r, err := c.DoAPIPOST(c.kvpath(prefix, id), utils.ToJSON(in)) // nolint:bodyclose
	if err != nil {
		return false, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	var out map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return false, model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}

	changed := out["changed"].(bool)

	return changed, model.BuildResponse(r), nil
}

func (c *ClientPP) KVGet(prefix, id string, ref interface{}) (*model.Response, error) {
	r, err := c.DoAPIGET(c.kvpath(prefix, id), "") // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}

	err = json.Unmarshal(buf, ref)
	if err != nil {
		return model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}

	return model.BuildResponse(r), nil
}

func (c *ClientPP) KVDelete(prefix, id string) (*model.Response, error) {
	r, err := c.DoAPIDELETE(c.kvpath(prefix, id)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) Subscribe(sub *apps.Subscription) (*model.Response, error) {
	data, err := json.Marshal(sub)
	if err != nil {
		return nil, err
	}
	r, err := c.DoAPIPOST(c.apipath(appspath.Subscribe), string(data)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) GetSubscriptions() ([]apps.Subscription, *model.Response, error) {
	r, err := c.DoAPIGET(c.apipath(appspath.Subscribe), "") // nolint:bodyclose
	if err != nil {
		return nil, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	var subs []apps.Subscription
	err = json.NewDecoder(r.Body).Decode(&subs)
	if err != nil {
		return nil, model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}

	return subs, model.BuildResponse(r), nil
}

func (c *ClientPP) Unsubscribe(sub *apps.Subscription) (*model.Response, error) {
	data, err := json.Marshal(sub)
	if err != nil {
		return nil, err
	}
	r, err := c.DoAPIPOST(c.apipath(appspath.Unsubscribe), string(data)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) StoreOAuth2App(oauth2App apps.OAuth2App) (*model.Response, error) {
	r, err := c.DoAPIPOST(c.apipath(appspath.OAuth2App), utils.ToJSON(oauth2App)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) StoreOAuth2User(ref interface{}) (*model.Response, error) {
	r, err := c.DoAPIPOST(c.apipath(appspath.OAuth2User), utils.ToJSON(ref)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) GetOAuth2User(ref interface{}) (*model.Response, error) {
	r, err := c.DoAPIGET(c.apipath(appspath.OAuth2User), "") // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	err = json.NewDecoder(r.Body).Decode(ref)
	if err != nil {
		return model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}

	return model.BuildResponse(r), nil
}

type UpdateAppListingRequest struct {
	// Manifest is the new app manifest to list.
	apps.Manifest

	// Replace causes the previously listed manifest to be dropped entirely.
	// When false, the deployment data from the new manifest will be combined
	// with the prerviously listed one.
	Replace bool

	// AddDeploys specifies which deployment types should be added to the
	// listing.
	AddDeploys apps.DeployTypes `json:"add_deploys,omitempty"`

	// RemoveDeploys specifies which deployment types should be removed from
	// the listing.
	RemoveDeploys apps.DeployTypes `json:"remove_deploys,omitempty"`
}

// UpdateAppListing adds a specified App manifest to the local store.
func (c *ClientPP) UpdateAppListing(req UpdateAppListingRequest) (*model.Response, error) {
	b, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	r, err := c.DoAPIPOST(c.apipath(appspath.UpdateAppListing), string(b)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

// InstallApp installs a app using a given manfest.
func (c *ClientPP) InstallApp(appID apps.AppID, deployType apps.DeployType) (*model.Response, error) {
	b, err := json.Marshal(apps.App{
		Manifest: apps.Manifest{
			AppID: appID,
		},
		DeployType: deployType,
	})
	if err != nil {
		return nil, err
	}
	r, err := c.DoAPIPOST(c.apipath(appspath.InstallApp), string(b)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) UninstallApp(appID apps.AppID) (*model.Response, error) {
	b, err := json.Marshal(apps.Manifest{
		AppID: appID,
	})
	if err != nil {
		return nil, err
	}
	r, err := c.DoAPIPOST(c.apipath(appspath.InstallApp), string(b)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) GetApp(appID apps.AppID) (*apps.App, *model.Response, error) {
	r, err := c.DoAPIGET(c.apipath(appspath.Apps)+"/"+string(appID), "") // nolint:bodyclose
	if err != nil {
		return nil, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	var app apps.App
	err = json.NewDecoder(r.Body).Decode(&app)
	if err != nil {
		return nil, model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}

	return &app, model.BuildResponse(r), nil
}

func (c *ClientPP) EnableApp(appID apps.AppID) (*model.Response, error) {
	b, err := json.Marshal(apps.Manifest{
		AppID: appID,
	})
	if err != nil {
		return nil, err
	}
	r, err := c.DoAPIPOST(c.apipath(appspath.EnableApp), string(b)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) DisableApp(appID apps.AppID) (*model.Response, error) {
	b, err := json.Marshal(apps.Manifest{
		AppID: appID,
	})
	if err != nil {
		return nil, err
	}
	r, err := c.DoAPIPOST(c.apipath(appspath.DisableApp), string(b)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) GetListedApps(filter string, includePlugins bool) ([]apps.ListedApp, *model.Response, error) {
	v := url.Values{}
	v.Add("filter", filter)
	if includePlugins {
		v.Add("include_plugins", "true")
	}
	r, err := c.DoAPIGET(c.apipath(appspath.Marketplace)+"?"+v.Encode(), "") // nolint:bodyclose
	if err != nil {
		return nil, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	listed := []apps.ListedApp{}
	err = json.NewDecoder(r.Body).Decode(&listed)
	if err != nil {
		return nil, model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}
	return listed, model.BuildResponse(r), nil
}

func (c *ClientPP) Call(creq apps.CallRequest) (*apps.CallResponse, *model.Response, error) {
	b, err := json.Marshal(&creq)
	if err != nil {
		return nil, nil, err
	}

	r, err := c.DoAPIPOST(c.apipath(appspath.Call), string(b)) // nolint:bodyclose
	if err != nil {
		return nil, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	var cresp apps.CallResponse
	err = json.NewDecoder(r.Body).Decode(&cresp)
	if err != nil {
		return nil, model.BuildResponse(r), errors.Wrap(err, "failed to decode response")
	}

	return &cresp, model.BuildResponse(r), nil
}

func (c *ClientPP) getPluginsRoute() string {
	return "/plugins"
}

func (c *ClientPP) GetPluginRoute(pluginID string) string {
	path := "/" + pluginID
	if c.fromPlugin {
		return path
	}

	return c.getPluginsRoute() + path
}

func (c *ClientPP) DoAPIGET(url string, etag string) (*http.Response, error) {
	return c.DoAPIRequest(http.MethodGet, c.URL+url, "", etag)
}

func (c *ClientPP) DoAPIPOST(url string, data string) (*http.Response, error) {
	return c.DoAPIRequest(http.MethodPost, c.URL+url, data, "")
}

func (c *ClientPP) DoAPIDELETE(url string) (*http.Response, error) {
	return c.DoAPIRequest(http.MethodDelete, c.URL+url, "", "")
}

func (c *ClientPP) DoAPIRequest(method, url, data, etag string) (*http.Response, error) {
	return c.doAPIRequestReader(method, url, strings.NewReader(data), etag)
}

func (c *ClientPP) doAPIRequestReader(method, url string, data io.Reader, etag string) (*http.Response, error) {
	rq, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	if len(etag) > 0 {
		rq.Header.Set(HeaderEtagClient, etag)
	}

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HeaderAuth, c.AuthType+" "+c.AuthToken)
	}

	if len(c.HTTPHeader) > 0 {
		for k, v := range c.HTTPHeader {
			rq.Header.Set(k, v)
		}
	}

	rp, err := c.HTTPClient.Do(rq)
	if err != nil {
		return rp, err
	}

	if rp.StatusCode == 304 {
		return rp, nil
	}

	if rp.StatusCode >= 300 {
		defer c.closeBody(rp)
		data, err := io.ReadAll(rp.Body)
		if err != nil {
			return rp, err
		}

		return rp, errors.New(strings.TrimSpace(string(data)))
	}

	return rp, nil
}

func (c *ClientPP) closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}
}

func (c *ClientPP) apipath(p string) string {
	return c.GetPluginRoute(AppsPluginName) + appspath.API + p
}

func (c *ClientPP) kvpath(prefix, id string) string {
	return c.apipath(path.Join("/kv", prefix, id))
}
