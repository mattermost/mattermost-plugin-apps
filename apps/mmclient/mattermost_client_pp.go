package mmclient

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	HeaderEtagClient = "If-None-Match"
	HeaderAuth       = "Authorization"

	AppsPluginName = "com.mattermost.apps"
)

// Paths for the REST APIs exposed by the Apps Plugin itself
const (
	// Top-level path
	PathAPI = "/api/v1"

	// Other sub-paths.
	PathKV          = "/kv"
	PathSubscribe   = "/subscribe"
	PathUnsubscribe = "/unsubscribe"

	PathApps      = "/apps"
	PathApp       = "/app"
	PathEnable    = "/enable"
	PathDisable   = "/disable"
	PathUninstall = "/uninstall"

	PathBotIDs      = "/bot-ids"
	PathOAuthAppIDs = "/oauth-app-ids"

	PathOAuth2App         = "/oauth2/app"
	PathOAuth2User        = "/oauth2/user"
	PathOAuth2CreateState = "/oauth2/create-state"
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

func (c *ClientPP) KVSet(id string, prefix string, in interface{}) (bool, *model.Response, error) {
	r, err := c.DoAPIPOST(c.kvpath(prefix, id), utils.ToJSON(in)) // nolint:bodyclose
	if err != nil {
		return false, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	var out map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return false, model.BuildResponse(r), err
	}

	changed := out["changed"].(bool)

	return changed, model.BuildResponse(r), nil
}

func (c *ClientPP) KVGet(id string, prefix string, ref interface{}) (*model.Response, error) {
	r, err := c.DoAPIGET(c.kvpath(prefix, id), "") // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	err = json.NewDecoder(r.Body).Decode(ref)
	if err != nil {
		return model.BuildResponse(r), err
	}

	return model.BuildResponse(r), nil
}

func (c *ClientPP) KVDelete(id string, prefix string) (*model.Response, error) {
	r, err := c.DoAPIDELETE(c.kvpath(prefix, id)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) Subscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response, error) {
	r, err := c.DoAPIPOST(c.apipath(PathSubscribe), request.ToJSON()) // nolint:bodyclose
	if err != nil {
		return nil, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	subResponse, err := apps.SubscriptionResponseFromJSON(r.Body)
	if err != nil {
		return nil, model.BuildResponse(r), err
	}

	return subResponse, model.BuildResponse(r), nil
}

func (c *ClientPP) Unsubscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response, error) {
	r, err := c.DoAPIPOST(c.apipath(PathUnsubscribe), request.ToJSON()) // nolint:bodyclose
	if err != nil {
		return nil, model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	subResponse, err := apps.SubscriptionResponseFromJSON(r.Body)
	if err != nil {
		return nil, model.BuildResponse(r), err
	}

	return subResponse, model.BuildResponse(r), nil
}

func (c *ClientPP) StoreOAuth2App(appID apps.AppID, clientID, clientSecret string) (*model.Response, error) {
	data := utils.ToJSON(apps.OAuth2App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	r, err := c.DoAPIPOST(c.apipath(PathOAuth2App)+"/"+string(appID), data) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) StoreOAuth2User(appID apps.AppID, ref interface{}) (*model.Response, error) {
	r, err := c.DoAPIPOST(c.apipath(PathOAuth2User)+"/"+string(appID), utils.ToJSON(ref)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) GetOAuth2User(appID apps.AppID, ref interface{}) (*model.Response, error) {
	r, err := c.DoAPIGET(c.apipath(PathOAuth2User)+"/"+string(appID), "") // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	err = json.NewDecoder(r.Body).Decode(ref)
	if err != nil {
		return model.BuildResponse(r), err
	}

	return model.BuildResponse(r), nil
}

// InstallApp installs a app using a given manfest.
func (c *ClientPP) InstallApp(m apps.Manifest) (*model.Response, error) {
	b, err := json.Marshal(&m)
	if err != nil {
		return nil, err
	}

	r, err := c.DoAPIPOST(c.apipath(PathApps), string(b)) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) UninstallApp(appID apps.AppID) (*model.Response, error) {
	r, err := c.DoAPIDELETE(c.apipath(PathApps) + "/" + string(appID) + PathUninstall) // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) GetApp(appID apps.AppID) (*apps.App, *model.Response, error) {
	r, err := c.DoAPIGET(c.apipath(PathApps)+"/"+string(appID), "") // nolint:bodyclose
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
	r, err := c.DoAPIPOST(c.apipath(PathApps)+"/"+string(appID)+PathEnable, "") // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
}

func (c *ClientPP) DisableApp(appID apps.AppID) (*model.Response, error) {
	r, err := c.DoAPIPOST(c.apipath(PathApps)+"/"+string(appID)+PathDisable, "") // nolint:bodyclose
	if err != nil {
		return model.BuildResponse(r), err
	}
	defer c.closeBody(r)

	return model.BuildResponse(r), nil
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

		return rp, errors.New((string(data)))
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
	return c.GetPluginRoute(AppsPluginName) + PathAPI + p
}

func (c *ClientPP) kvpath(prefix, id string) string {
	return c.apipath(path.Join("/kv", prefix, id))
}
