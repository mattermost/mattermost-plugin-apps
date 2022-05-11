package httpin

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/gateway"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/restapi"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service interface {
	ServePluginHTTP(c *plugin.Context, w http.ResponseWriter, req *http.Request)
}

var _ Service = (*handler.Handler)(nil)
var _ http.Handler = (*handler.Handler)(nil)

func NewService(proxy proxy.Service, appservices appservices.Service, conf config.Service, log utils.Logger) *handler.Handler {
	h := handler.NewHandler(proxy, conf, log)

	gateway.Init(h)
	restapi.Init(h, appservices)

	return h
}
