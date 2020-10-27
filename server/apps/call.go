package apps

import "github.com/mattermost/mattermost-plugin-apps/server/api"

func (s *service) Call(call *api.Call) (*api.CallResponse, error) {
	var err error
	req := *call
	// TODO Expand using the App's bot credentials!
	req.Context, err = s.newExpander(call.Context).Expand(call.Expand)
	if err != nil {
		return nil, err
	}
	req.Expand = nil
	req.FormURL = ""

	return s.Client.PostCall(call)
}
