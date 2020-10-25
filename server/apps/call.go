package apps

import "github.com/mattermost/mattermost-plugin-apps/server/api"

func (s *service) Call(call *api.Call) (*api.CallResponse, error) {
	// TODO <><> get a cached function
	f, err := s.Client.GetFunction(call)
	if err != nil {
		return nil, err
	}

	req := *call
	// TODO Expand using the App's bot credentials!
	req.Context, err = s.newExpander(call.Context).Expand(f.Expand)
	if err != nil {
		return nil, err
	}
	req.URL = ""

	return s.Client.PostFunction(&req)
}
