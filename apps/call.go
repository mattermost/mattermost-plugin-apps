// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// Call defines a way to invoke an App's function. Calls are used to fetch App's
// bindings, to process notifications, and to respond to user input from forms,
// bindings and command line.
//
// IMPORTANT: update UnmarshalJSON if this struct changes.
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

func NewCall(path string) *Call {
	c := Call{
		Path: path,
	}
	return &c
}

func (c Call) WithExpand(expand Expand) *Call {
	c.Expand = &expand
	return &c
}

func (c Call) WithState(state interface{}) *Call {
	c.State = state
	return &c
}

func (c Call) WithLocale() *Call {
	if c.Expand == nil {
		c.Expand = &Expand{}
	}
	c.Expand.Locale = ExpandAll
	return &c
}

func (c *Call) WithDefault(def Call) Call {
	if c == nil {
		return *def.PartialCopy()
	}
	clone := c.PartialCopy()

	if clone.Path == "" {
		clone.Path = def.Path
	}
	if clone.Expand == nil {
		clone.Expand = def.Expand
	}
	if clone.State == nil {
		clone.State = def.State
	}
	return *clone
}

func (c *Call) PartialCopy() *Call {
	if c == nil {
		return nil
	}

	clone := *c
	if clone.Expand != nil {
		cloneExpand := *clone.Expand
		clone.Expand = &cloneExpand
	}

	// Only know how to clone map values for State.
	if state, ok := clone.State.(map[string]interface{}); ok {
		cloneState := map[string]interface{}{}
		for k, v := range state {
			cloneState[k] = v
		}
		clone.State = cloneState
	}
	if state, ok := clone.State.(map[string]string); ok {
		cloneState := map[string]string{}
		for k, v := range state {
			cloneState[k] = v
		}
		clone.State = cloneState
	}
	return &clone
}

func (c Call) String() string {
	s := c.Path
	if c.Expand != nil {
		s += fmt.Sprintf(", expand: %v", c.Expand.String())
	}
	if c.State != nil {
		s += fmt.Sprintf(", state: %v", utils.LogDigest(c.State))
	}
	return s
}

func (c Call) Loggable() []interface{} {
	props := []interface{}{"call_path", c.Path}
	if c.Expand != nil {
		props = append(props, "call_expand", c.Expand.String())
	}
	if c.State != nil {
		props = append(props, "call_state", utils.LogDigest(c.State))
	}
	return props
}
