package mmclient

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	HEADER_ETAG_CLIENT = "If-None-Match"
	HEADER_AUTH        = "Authorization"

	API_PATH_PP       = "/api/v1"
	API_URL_SUFFIX_V4 = "/api/v4"
	API_URL_SUFFIX    = API_URL_SUFFIX_V4
	APPS_PLUGIN_NAME  = "com.mattermost.apps"
)

type ClientPP struct {
	Url        string       // The location of the server, for example  "http://localhost:8065"
	ApiUrl     string       // The api location of the server, for example "http://localhost:8065/api/v4"
	HttpClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	HttpHeader map[string]string // Headers to be copied over for each request

	// TrueString is the string value sent to the server for true boolean query parameters.
	trueString string

	// FalseString is the string value sent to the server for false boolean query parameters.
	falseString string
}

func NewAPIClientPP(url string) *ClientPP {
	url = strings.TrimRight(url, "/")
	return &ClientPP{url, url, &http.Client{}, "", "", map[string]string{}, "", ""}
}

func (c *ClientPP) KVSet(id string, prefix string, in map[string]interface{}) (map[string]interface{}, *model.Response) {
	query := fmt.Sprintf("%v/kv/%v?prefix=%v", API_PATH_PP, id, prefix)
	r, appErr := c.DoApiPost(c.GetPluginRoute(APPS_PLUGIN_NAME)+query, StringInterfaceToJSON(in))

	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return StringInterfaceFromJSON(r.Body), model.BuildResponse(r)
}

func (c *ClientPP) KVGet(id string, prefix string) (map[string]interface{}, *model.Response) {
	query := fmt.Sprintf("%v/kv/%v?prefix=%v", API_PATH_PP, id, prefix)
	r, appErr := c.DoApiGet(c.GetPluginRoute(APPS_PLUGIN_NAME)+query, "")
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return StringInterfaceFromJSON(r.Body), model.BuildResponse(r)
}

func (c *ClientPP) KVDelete(id string, prefix string) (bool, *model.Response) {
	query := fmt.Sprintf("%v/kv/%v?prefix=%v", API_PATH_PP, id, prefix)
	r, appErr := c.DoApiDelete(c.GetPluginRoute(APPS_PLUGIN_NAME) + query)
	if appErr != nil {
		return false, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return model.CheckStatusOK(r), model.BuildResponse(r)
}

func (c *ClientPP) Subscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response) {
	r, appErr := c.DoApiPost(c.GetPluginRoute(APPS_PLUGIN_NAME)+API_PATH_PP+"/subscribe", request.ToJson())
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	subResponse := apps.SubscriptionResponseFromJson(r.Body)
	return subResponse, model.BuildResponse(r)
}

func (c *ClientPP) Unsubscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response) {
	r, appErr := c.DoApiPost(c.GetPluginRoute(APPS_PLUGIN_NAME)+API_PATH_PP+"/unsubscribe", request.ToJson())
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	subResponse := apps.SubscriptionResponseFromJson(r.Body)
	return subResponse, model.BuildResponse(r)
}

func (c *ClientPP) GetPluginsRoute() string {
	return "/plugins"
}

func (c *ClientPP) GetPluginRoute(pluginID string) string {
	return fmt.Sprintf(c.GetPluginsRoute()+"/%v", pluginID)
}

func (c *ClientPP) DoApiGet(url string, etag string) (*http.Response, *model.AppError) {
	return c.DoApiRequest(http.MethodGet, c.ApiUrl+url, "", etag)
}

func (c *ClientPP) DoApiPost(url string, data string) (*http.Response, *model.AppError) {
	return c.DoApiRequest(http.MethodPost, c.ApiUrl+url, data, "")
}

func (c *ClientPP) DoApiDelete(url string) (*http.Response, *model.AppError) {
	return c.DoApiRequest(http.MethodDelete, c.ApiUrl+url, "", "")
}

func (c *ClientPP) DoApiRequest(method, url, data, etag string) (*http.Response, *model.AppError) {
	return c.doApiRequestReader(method, url, strings.NewReader(data), etag)
}

func (c *ClientPP) doApiRequestReader(method, url string, data io.Reader, etag string) (*http.Response, *model.AppError) {
	rq, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, model.NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	if len(etag) > 0 {
		rq.Header.Set(HEADER_ETAG_CLIENT, etag)
	}

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if c.HttpHeader != nil && len(c.HttpHeader) > 0 {
		for k, v := range c.HttpHeader {
			rq.Header.Set(k, v)
		}
	}

	rp, err := c.HttpClient.Do(rq)

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

func StringInterfaceToJSON(objmap map[string]interface{}) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func StringInterfaceFromJSON(data io.Reader) map[string]interface{} {
	decoder := json.NewDecoder(data)

	var objmap map[string]interface{}
	if err := decoder.Decode(&objmap); err != nil {
		return make(map[string]interface{})
	}
	return objmap
}

func (c *ClientPP) closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}
}
