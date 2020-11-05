package apps

import "github.com/mattermost/mattermost-plugin-apps/server/api"

func (s *service) Call(call *api.Call) (*api.CallResponse, error) {
	// TODO <><> get a cached function
	f, err := s.Client.GetFunctionMeta(call)
	if err != nil {
		return nil, err
	}

	if call.AsModal {
		return &api.CallResponse{
			Type: api.CallResponseTypeOK,
			Form: f.Form,
		}, nil
	}

	req := *call
	// TODO Expand using the App's bot credentials!
	req.Context, err = s.newExpander(call.Context).Expand(f.Expand)
	if err != nil {
		return nil, err
	}

	return s.Client.PostFunction(&req)
}
