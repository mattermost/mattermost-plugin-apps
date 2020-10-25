package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *service) GetBindings(cc *api.Context) ([]*api.Binding, error) {
	return nil, nil
}
