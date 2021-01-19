package modelapps

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	HEADER_ETAG_CLIENT        = "If-None-Match"
	HEADER_AUTH               = "Authorization"

	API_URL_SUFFIX_V1 = "/api/v1"
	API_URL_SUFFIX    = API_URL_SUFFIX_V1
	APPS_PLUGIN_NAME = "com.mattermost.apps"
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
	return &ClientPP{url, url + API_URL_SUFFIX, &http.Client{}, "", "", map[string]string{}, "", ""}
}

func (c *ClientPP) Subscribe(request *Subscription) (*model.PluginsResponse, *model.Response) {
       r, appErr := c.DoApiPost(c.GetPluginRoute(APPS_PLUGIN_NAME)+"/subscribe", request.ToJson())
       if appErr != nil {
               return nil, model.BuildErrorResponse(r, appErr)
       }
       defer closeBody(r)
       return model.PluginsResponseFromJson(r.Body), model.BuildResponse(r)
}

//TODO this is right now a HTTP DELETE op - which does not accept a payload - needs to be converted to a POST(?)
func (c *ClientPP) Unsubscribe(request *Subscription) (bool, *model.Response) {
	   r, appErr := c.DoApiPost(c.GetPluginRoute(APPS_PLUGIN_NAME)+"/delete", request.ToJson())
       if appErr != nil {
               return false, model.BuildErrorResponse(r, appErr)
       }
       defer closeBody(r)
       return model.CheckStatusOK(r), model.BuildResponse(r)
}

func (c *ClientPP) GetPluginsRoute() string {
	return "/plugins"
}

func (c *ClientPP) GetPluginRoute(pluginId string) string {
	return fmt.Sprintf(c.GetPluginsRoute()+"/%v", pluginId)
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
		defer closeBody(rp)
		return rp, model.AppErrorFromJson(rp.Body)
	}

	return rp, nil
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}
}