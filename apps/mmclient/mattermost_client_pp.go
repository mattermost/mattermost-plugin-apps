package mmclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

const (
	HeaderEtagClient = "If-None-Match"
	HeaderAuth       = "Authorization"

	AppsPluginName = "com.mattermost.apps"
)

// Paths for the REST APIs exposed by the proxy plugin itself
const (
	// Top-level path
	PathAPI = "/api/v1"

	// Other sub-paths.
	PathKV          = "/kv"
	PathSubscribe   = "/subscribe"
	PathUnsubscribe = "/unsubscribe"

	PathOAuth2App         = "/oauth2/app"
	PathOAuth2User        = "/oauth2/user"
	PathOAuth2CreateState = "/oauth2/create-state"
)

type ClientPP struct {
	URL        string       // The location of the server, for example  "http://localhost:8065"
	APIURL     string       // The api location of the server, for example "http://localhost:8065/api/v4"
	HTTPClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	HTTPHeader map[string]string // Headers to be copied over for each request

	// TrueString is the string value sent to the server for true boolean query parameters.
	trueString string

	// FalseString is the string value sent to the server for false boolean query parameters.
	falseString string
}

func NewAPIClientPP(url string) *ClientPP {
	url = strings.TrimRight(url, "/")
	return &ClientPP{url, url, &http.Client{}, "", "", map[string]string{}, "", ""}
}

func (c *ClientPP) SetOAuthToken(token string) {
	c.AuthToken = token
	c.AuthType = model.HEADER_BEARER
}

func (c *ClientPP) KVSet(id string, prefix string, in interface{}) (interface{}, *model.Response) {
	r, appErr := c.DoAPIPOST(c.kvpath(prefix, id), utils.ToJSON(in)) // nolint:bodyclose
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return interfaceFromJSON(r.Body), model.BuildResponse(r)
}

func (c *ClientPP) KVGet(id string, prefix string, ref interface{}) *model.Response {
	r, appErr := c.DoAPIGET(c.kvpath(prefix, id), "") // nolint:bodyclose
	if appErr != nil {
		return model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	err := json.NewDecoder(r.Body).Decode(ref)
	if err != nil {
		return model.BuildErrorResponse(r, model.NewAppError("KVGet", "", nil, err.Error(), http.StatusInternalServerError))
	}
	return model.BuildResponse(r)
}

func (c *ClientPP) KVDelete(id string, prefix string) (bool, *model.Response) {
	r, appErr := c.DoAPIDELETE(c.kvpath(prefix, id)) // nolint:bodyclose
	if appErr != nil {
		return false, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return model.CheckStatusOK(r), model.BuildResponse(r)
}

func (c *ClientPP) Subscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response) {
	r, appErr := c.DoAPIPOST(c.apipath(PathSubscribe), request.ToJSON()) // nolint:bodyclose
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	subResponse := apps.SubscriptionResponseFromJSON(r.Body)
	return subResponse, model.BuildResponse(r)
}

func (c *ClientPP) Unsubscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response) {
	r, appErr := c.DoAPIPOST(c.apipath(PathUnsubscribe), request.ToJSON()) // nolint:bodyclose
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	subResponse := apps.SubscriptionResponseFromJSON(r.Body)
	return subResponse, model.BuildResponse(r)
}

func (c *ClientPP) StoreOAuth2App(appID apps.AppID, clientID, clientSecret string) *model.Response {
	data := utils.ToJSON(apps.OAuth2App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	r, appErr := c.DoAPIPOST(c.apipath(PathOAuth2App)+"/"+string(appID), data) // nolint:bodyclose
	if appErr != nil {
		return model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return model.BuildResponse(r)
}

func (c *ClientPP) StoreOAuth2User(appID apps.AppID, ref interface{}) *model.Response {
	r, appErr := c.DoAPIPOST(c.apipath(PathOAuth2User)+"/"+string(appID), utils.ToJSON(ref)) // nolint:bodyclose
	if appErr != nil {
		return model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return model.BuildResponse(r)
}

func (c *ClientPP) GetOAuth2User(appID apps.AppID, ref interface{}) *model.Response {
	r, appErr := c.DoAPIGET(c.apipath(PathOAuth2User)+"/"+string(appID), "") // nolint:bodyclose
	if appErr != nil {
		return model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	err := json.NewDecoder(r.Body).Decode(ref)
	if err != nil {
		return model.BuildErrorResponse(r, model.NewAppError("GetOAuth2User", "", nil, err.Error(), http.StatusInternalServerError))
	}
	return model.BuildResponse(r)
}

func (c *ClientPP) GetPluginsRoute() string {
	return "/plugins"
}

func (c *ClientPP) GetPluginRoute(pluginID string) string {
	return fmt.Sprintf(c.GetPluginsRoute()+"/%v", pluginID)
}

func (c *ClientPP) DoAPIGET(url string, etag string) (*http.Response, *model.AppError) {
	return c.DoAPIRequest(http.MethodGet, c.APIURL+url, "", etag)
}

func (c *ClientPP) DoAPIPOST(url string, data string) (*http.Response, *model.AppError) {
	return c.DoAPIRequest(http.MethodPost, c.APIURL+url, data, "")
}

func (c *ClientPP) DoAPIDELETE(url string) (*http.Response, *model.AppError) {
	return c.DoAPIRequest(http.MethodDelete, c.APIURL+url, "", "")
}

func (c *ClientPP) DoAPIRequest(method, url, data, etag string) (*http.Response, *model.AppError) {
	return c.doAPIRequestReader(method, url, strings.NewReader(data), etag)
}

func (c *ClientPP) doAPIRequestReader(method, url string, data io.Reader, etag string) (*http.Response, *model.AppError) {
	rq, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, model.NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)
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

	if err != nil || rp == nil {
		return nil, model.NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0)
	}

	if rp.StatusCode == 304 {
		return rp, nil
	}

	if rp.StatusCode >= 300 {
		defer c.closeBody(rp)
		return rp, model.AppErrorFromJson(rp.Body)
	}

	return rp, nil
}

func interfaceFromJSON(data io.Reader) interface{} {
	decoder := json.NewDecoder(data)

	var objmap interface{}
	if err := decoder.Decode(&objmap); err != nil {
		return ""
	}
	return objmap
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
