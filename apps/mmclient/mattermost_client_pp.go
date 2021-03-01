package mmclient

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const (
	HeaderEtagClient = "If-None-Match"
	HeaderAuth       = "Authorization"

	APIPathPP      = "/api/v1"
	APIUrlSuffixV4 = "/api/v4"
	APIUrlSuffix   = APIUrlSuffixV4
	AppsPluginName = "com.mattermost.apps"
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

func (c *ClientPP) KVSet(id string, prefix string, in map[string]interface{}) (map[string]interface{}, *model.Response) {
	query := fmt.Sprintf("%v/kv/%v?prefix=%v", APIPathPP, id, prefix)
	r, appErr := c.DoAPIPOST(c.GetPluginRoute(AppsPluginName)+query, StringInterfaceToJSON(in)) // nolint:bodyclose

	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return StringInterfaceFromJSON(r.Body), model.BuildResponse(r)
}

func (c *ClientPP) KVGet(id string, prefix string) (map[string]interface{}, *model.Response) {
	query := fmt.Sprintf("%v/kv/%v?prefix=%v", APIPathPP, id, prefix)
	r, appErr := c.DoAPIGET(c.GetPluginRoute(AppsPluginName)+query, "") // nolint:bodyclose
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return StringInterfaceFromJSON(r.Body), model.BuildResponse(r)
}

func (c *ClientPP) KVDelete(id string, prefix string) (bool, *model.Response) {
	query := fmt.Sprintf("%v/kv/%v?prefix=%v", APIPathPP, id, prefix)
	r, appErr := c.DoAPIDELETE(c.GetPluginRoute(AppsPluginName) + query) // nolint:bodyclose
	if appErr != nil {
		return false, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)
	return model.CheckStatusOK(r), model.BuildResponse(r)
}

func (c *ClientPP) Subscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response) {
	r, appErr := c.DoAPIPOST(c.GetPluginRoute(AppsPluginName)+APIPathPP+"/subscribe", request.ToJSON()) // nolint:bodyclose
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	subResponse := apps.SubscriptionResponseFromJSON(r.Body)
	return subResponse, model.BuildResponse(r)
}

func (c *ClientPP) Unsubscribe(request *apps.Subscription) (*apps.SubscriptionResponse, *model.Response) {
	r, appErr := c.DoAPIPOST(c.GetPluginRoute(AppsPluginName)+APIPathPP+"/unsubscribe", request.ToJSON()) // nolint:bodyclose
	if appErr != nil {
		return nil, model.BuildErrorResponse(r, appErr)
	}
	defer c.closeBody(r)

	subResponse := apps.SubscriptionResponseFromJSON(r.Body)
	return subResponse, model.BuildResponse(r)
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
