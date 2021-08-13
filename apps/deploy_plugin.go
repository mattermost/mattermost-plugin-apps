package apps

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const PluginAppPath = "/app"

// Plugin contains metadata for an app that is implemented and is deployed and
// accessed as a local Plugin. The JSON name `plugin` must match the type.
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