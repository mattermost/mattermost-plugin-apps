// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// CallRequest envelops all requests sent to Apps.
type CallRequest struct {
	// A copy of the Call struct that originated the request. Path and State are
	// of significance.
	Call

	// Values are all values entered by the user.
	Values map[string]interface{} `json:"values,omitempty"`

	// Context of execution, see the Context type for more information.
	Context Context `json:"context,omitempty"`

	// In case the request came from the command line, the raw text of the
	// command, as submitted by the user.
	RawCommand string `json:"raw_command,omitempty"`

	// In the case of a form refresh call, the field of the select value chosen.
	// In the case of a lookup call, the field that the user is using for autocomplete.
	SelectedField string `json:"selected_field,omitempty"`

	// In the case of a lookup call, the query the user has typed in for autocomplete.
	Query string `json:"query,omitempty"`
}

// UnmarshalJSON has to be defined since Call is embedded anonymously, and
// CallRequest inherits its UnmarshalJSON unless it defines its own.
func (creq *CallRequest) UnmarshalJSON(data []byte) error {
	// Unmarshal the Call first
	call := Call{}
	err := json.Unmarshal(data, &call)
	if err != nil {
		return err
	}

	// Need a type that is just like CallRequest, but without Call to avoid
	// recursion.
	structValue := struct {
		Values        map[string]interface{} `json:"values,omitempty"`
		Context       Context                `json:"context,omitempty"`
		RawCommand    string                 `json:"raw_command,omitempty"`
		SelectedField string                 `json:"selected_field,omitempty"`
		Query         string                 `json:"query,omitempty"`
	}{}
	err = json.Unmarshal(data, &structValue)
	if err != nil {
		return err
	}

	*creq = CallRequest{
		Call:          call,
		Values:        structValue.Values,
		Context:       structValue.Context,
		RawCommand:    structValue.RawCommand,
		SelectedField: structValue.SelectedField,
		Query:         structValue.Query,
	}
	return nil
}

func CallRequestFromJSON(data []byte) (*CallRequest, error) {
	c := CallRequest{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CallRequestFromJSONReader(in io.Reader) (*CallRequest, error) {
	c := CallRequest{}
	err := json.NewDecoder(in).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (creq *CallRequest) GetValue(name, defaultValue string) string {
	if len(creq.Values) == 0 {
		return defaultValue
	}

	s, ok := creq.Values[name].(string)
	if ok && s != "" {
		return s
	}

	opt, ok := creq.Values[name].(map[string]interface{})
	if ok {
		if v, ok2 := opt["value"].(string); ok2 {
			return v
		}
	}

	return defaultValue
}

func (creq *CallRequest) BoolValue(name string) bool {
	if len(creq.Values) == 0 {
		return false
	}

	isBool := func(v interface{}) (bool, bool) {
		if b, ok := v.(bool); ok {
			return b, true
		}
		if b, ok := creq.Values[name].(string); ok {
			switch b {
			case "true":
				return true, true
			case "false":
				return false, true
			}
		}
		return false, false
	}

	if b, ok := isBool(creq.Values[name]); ok {
		return b
	}

	opt, ok := creq.Values[name].(map[string]interface{})
	if ok {
		if v, ok2 := isBool(opt["value"]); ok2 {
			return v
		}
	}

	return false
}

func (creq CallRequest) String() string {
	s := fmt.Sprintf("call: %s, context: %s", creq.Call.String(), creq.Context.String())
	if len(creq.Values) > 0 {
		s += ", values: " + utils.LogDigest(creq.Values)
	}
	if creq.Query != "" {
		s += ", query: " + creq.Query
	}
	return s
}

func (creq CallRequest) Loggable() []interface{} {
	props := append([]interface{}{}, creq.Call, creq.Context)
	if len(creq.Values) > 0 {
		props = append(props, "values", utils.LogDigest(creq.Values))
	}
	if creq.Query != "" {
		props = append(props, "query", creq.Query)
	}
	return props
}
