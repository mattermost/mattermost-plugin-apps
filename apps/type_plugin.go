package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

const PluginAppPath = "/app"

type Plugin struct {
	// PluginID is the ID of the plugin, which manages the app, if there is one.
	PluginID string `json:"plugin_id,omitempty"`
}

func (p *Plugin) IsValid() error {
	if p == nil {
		return nil
	}
	if p.PluginID == "" {
		return utils.NewInvalidError(errors.New("plugin_id must be set for plugin apps"))
	}
	return nil
}
