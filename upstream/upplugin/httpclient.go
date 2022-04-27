package upplugin

import (
	"net/http"

	"github.com/pkg/errors"
)

type PluginHTTPAPI interface {
	HTTP(*http.Request) *http.Response
}

type pluginAPIRoundTripper struct {
	client PluginHTTPAPI
}

func (p *pluginAPIRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := p.client.HTTP(req)
	if resp == nil {
		return nil, errors.Errorf("Failed to make interplugin request")
	}

	return resp, nil
}

func MakePluginHTTPClient(api PluginHTTPAPI) http.Client {
	httpClient := http.Client{}
	httpClient.Transport = &pluginAPIRoundTripper{api}

	return httpClient
}
