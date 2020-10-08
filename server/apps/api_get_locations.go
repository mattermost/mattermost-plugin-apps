package apps

import "github.com/pkg/errors"

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *Service) GetLocations(userID, channelID string) ([]LocationInt, error) {
	ids, err := s.Registry.GetAllAppIDs()
	if err != nil {
		return nil, errors.Wrap(err, "error getting all app IDs")
	}

	allLocations := []LocationInt{}
	for _, id := range ids {
		locations, err := s.Client.GetLocationsFromApp(id, userID, channelID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get single location")
		}
		allLocations = append(allLocations, locations...)
	}

	return allLocations, nil
}
