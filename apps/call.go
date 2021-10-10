// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
)

// Call defines a way to invoke an App's function. Calls are used to fetch App's
// bindings, to process notifications, and to respond to user input from forms,
// bindings and command line.
type Call struct {
	// The path of the Call. For HTTP apps, the path is appended to the app's
	// RootURL. For AWS Lambda apps, it is mapped to the appropriate Lambda name
	// to invoke, and then passed in the call request.
	Path string `json:"path,omitempty"`

	// Expand specifies what extended data should be provided to the function in
	// each request's Context. It may be various auth tokens, configuration
	// data, or details of Mattermost entities such as the acting user, current
	// team and channel, etc.
	Expand *Expand `json:"expand,omitempty"`

	// Custom data that will be passed to the function in JSON, "as is".
	State interface{} `json:"state,omitempty"`
}

func (c *Call) UnmarshalJSON(data []byte) error {
	stringValue := ""
	err := json.Unmarshal(data, &stringValue)
	if err == nil {
		*c = Call{
			Path: stringValue,
		}
		return nil
	}

	// Need a type that is just like Call, but without UnmarshalJSON
	structValue := struct {
		Path   string      `json:"path,omitempty"`
		Expand *Expand     `json:"expand,omitempty"`
		State  interface{} `json:"state,omitempty"`
	}{}
	err = json.Unmarshal(data, &structValue)
	if err != nil {
		return err
	}

	*c = Call{
		Path:   structValue.Path,
		Expand: structValue.Expand,
		State:  structValue.State,
	}
	return nil
}

func NewCall(url string) *Call {
	return &Call{
		Path: url,
	}
}

func (c *Call) WithState(state interface{}) *Call {
	clone := Call{}
	if c != nil {
		clone = *c
	}
	clone.State = state
	return &clone
}

func (c *Call) WithDefault(def Call) *Call {
	if c == nil {
		return &def
	}
	clone := *c

	if clone.Path == "" {
		clone.Path = def.Path
	}
	if clone.Expand == nil {
		clone.Expand = def.Expand
	}
	if clone.State == nil {
		clone.State = def.State
	}
	return &clone
}
