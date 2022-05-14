// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestLoggable(t *testing.T) {
	var _ utils.HasLoggable = Call{}
	var _ utils.HasLoggable = CallRequest{}
	var _ utils.HasLoggable = CallResponse{}
	var _ utils.HasLoggable = Context{}

	var simpleContext = Context{
		ExpandedContext: ExpandedContext{
			BotUserID:      "id_of_bot_user",
			BotAccessToken: "bot_user_access_tokenXYZ",
		},
	}
	var simpleCall = Call{
		Path: "/some-path",
	}
	var fullCall = Call{
		Path: "/some-path",
		State: map[string]interface{}{
			"key1": "confidential1",
			"key2": "confidential2",
		},
		Expand: &Expand{
			User:                  ExpandAll,
			Channel:               ExpandSummary,
			ActingUserAccessToken: ExpandAll,
			OAuth2App:             ExpandAll,
		},
	}
	var simpleCallRequest = CallRequest{
		Call:    simpleCall,
		Context: simpleContext,
	}
	var fullCallRequest = CallRequest{
		Call:    fullCall,
		Context: simpleContext,
		Values: map[string]interface{}{
			"vkey1": "confidential1",
			"vkey2": "confidential2",
		},
	}
	var testData = map[string]interface{}{
		"A": "test",
		"B": 99,
	}
	var testForm = Form{
		Title: "name",
		Fields: []Field{
			{
				Name: "f1",
			},
			{
				Name: "f2",
			},
		},
		Submit: &simpleCall,
	}

	for name, test := range map[string]struct {
		In             interface{}
		ExpectedProps  []interface{}
		ExpectedString string
	}{
		"Context": {
			In: simpleContext,
			ExpectedProps: []interface{}{
				"is_not_submit", "true",
				"bot_user_id", "id_of_bot_user",
				"bot_access_token", "***nXYZ",
			},
			ExpectedString: "bot_access_token: ***nXYZ, bot_user_id: id_of_bot_user, is_not_submit: true",
		},
		"Call simple": {
			In: simpleCall,
			ExpectedProps: []interface{}{
				"call_path", "/some-path",
			},
			ExpectedString: "/some-path",
		},
		"Call full": {
			In: fullCall,
			ExpectedProps: []interface{}{
				"call_path", "/some-path",
				"call_expand", "acting_user_access_token:all,channel:summary,oauth2_app:all,user:all",
				"call_state", "key1,key2",
			},
			ExpectedString: "/some-path, expand: acting_user_access_token:all,channel:summary,oauth2_app:all,user:all, state: key1,key2",
		},
		"CallRequest simple": {
			In:             simpleCallRequest,
			ExpectedProps:  []interface{}{simpleCall, simpleContext},
			ExpectedString: "call: /some-path, context: bot_access_token: ***nXYZ, bot_user_id: id_of_bot_user, is_not_submit: true",
		},
		"CallRequest full": {
			In:             fullCallRequest,
			ExpectedProps:  []interface{}{fullCall, simpleContext, "values", "vkey1,vkey2"},
			ExpectedString: "call: /some-path, expand: acting_user_access_token:all,channel:summary,oauth2_app:all,user:all, state: key1,key2, context: bot_access_token: ***nXYZ, bot_user_id: id_of_bot_user, is_not_submit: true, values: vkey1,vkey2",
		},
		"CallResponse text": {
			In:             NewTextResponse("test"),
			ExpectedProps:  []interface{}{"response_type", "ok", "response_text", "test"},
			ExpectedString: "OK: test",
		},
		"CallResponse JSON data": {
			In:             NewDataResponse(testData),
			ExpectedProps:  []interface{}{"response_type", "ok", "response_data", "not shown"},
			ExpectedString: "OK: data type map[string]interface {}, value: map[A:test B:99]",
		},
		"CallResponse byte data": {
			In:             NewDataResponse([]byte("12345")),
			ExpectedProps:  []interface{}{"response_type", "ok", "response_data", "not shown"},
			ExpectedString: "OK: data type []uint8, value: [49 50 51 52 53]",
		},
		"CallResponse text data": {
			In:             NewDataResponse("12345"),
			ExpectedProps:  []interface{}{"response_type", "ok", "response_data", "not shown"},
			ExpectedString: "OK: data type string, value: 12345",
		},
		"CallResponse form": {
			In:             NewFormResponse(testForm),
			ExpectedProps:  []interface{}{"response_type", "form", "response_form", "not shown"},
			ExpectedString: `Form: not shown`,
		},
		"CallResponse navigate": {
			In: CallResponse{
				Type:               CallResponseTypeNavigate,
				NavigateToURL:      "http://x.y.z",
				UseExternalBrowser: true,
			},
			ExpectedProps:  []interface{}{"response_type", "navigate", "response_url", "http://x.y.z", "use_external_browser", true},
			ExpectedString: `Navigate to: "http://x.y.z", using external browser`,
		},
		"CallResponse call": {
			In: CallResponse{
				Type: CallResponseTypeCall,
				Call: &fullCall,
			},
			ExpectedProps:  []interface{}{"response_type", "call", "response_call", "/some-path, expand: acting_user_access_token:all,channel:summary,oauth2_app:all,user:all, state: key1,key2"},
			ExpectedString: `Call: /some-path, expand: acting_user_access_token:all,channel:summary,oauth2_app:all,user:all, state: key1,key2`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.ExpectedProps != nil {
				lp, ok := test.In.(utils.HasLoggable)
				require.True(t, ok)
				require.EqualValues(t, test.ExpectedProps, lp.Loggable())
			}

			if test.ExpectedString != "" {
				s, ok := test.In.(fmt.Stringer)
				require.True(t, ok)
				require.Equal(t, test.ExpectedString, s.String())
			}
		})
	}
}
