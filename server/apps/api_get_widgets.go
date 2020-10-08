package apps

import "github.com/pkg/errors"

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *Service) GetWidgets(userID, channelID string) ([]LocationInt, error) {
	registers, err := s.Registry.GetAllLocations()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get location registers")
	}

	locations := []LocationInt{}
	for _, register := range registers {
		location, err := s.Client.GetLocation(register, userID, channelID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get single location")
		}
		locations = append(locations, location)
	}

	return locations, nil
}
