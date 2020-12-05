// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func DecodeBindingsResponse(in io.Reader) ([]*api.Binding, error) {
	cr := api.CallResponse{
		Data: &[]*api.Binding{},
	}
	err := json.NewDecoder(in).Decode(&cr)
	if err != nil {
		return nil, err
	}

	bindings, ok := cr.Data.(*[]*api.Binding)
	if !ok {
		return nil, errors.New("failed to decode bindings")
	}
	return *bindings, nil
}
