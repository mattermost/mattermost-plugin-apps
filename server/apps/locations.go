package apps

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *service) GetLocations(userID, channelID string) ([]api.LocationInt, error) {
	ids, err := s.Store.ListApps()
	if err != nil {
		return nil, errors.Wrap(err, "error getting all app IDs")
	}

	allLocations := []api.LocationInt{}
	for _, id := range ids {
		locations, err := s.Client.GetLocations(id, userID, channelID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get single location")
		}
		allLocations = append(allLocations, locations...)
	}

	return allLocations, nil
}
