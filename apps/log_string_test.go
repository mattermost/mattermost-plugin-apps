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
		ActingUserID: "id_of_acting_user",
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
	// TODO <>/<> Add test cases for CallResponse

	for name, test := range map[string]struct {
		In             interface{}
		ExpectedProps  []interface{}
		ExpectedString string
	}{
		"simple Context": {
			In: simpleContext,
			ExpectedProps: []interface{}{
				"is_not_submit", "true",
				"bot_user_id", "id_of_bot_user",
				"bot_access_token", "***nXYZ",
			},
			ExpectedString: "bot_access_token: ***nXYZ, bot_user_id: id_of_bot_user, is_not_submit: true",
		},
		"simple Call": {
			In: simpleCall,
			ExpectedProps: []interface{}{
				"call_path", "/some-path",
			},
			ExpectedString: "/some-path",
		},
		"full Call": {
			In: fullCall,
			ExpectedProps: []interface{}{
				"call_path", "/some-path",
				"call_expand", "acting_user_access_token:all,channel:summary,oauth2_app:all,user:all",
				"call_state", "key1,key2",
			},
			ExpectedString: "/some-path, expand: acting_user_access_token:all,channel:summary,oauth2_app:all,user:all, state: key1,key2",
		},
		"simple CallRequest": {
			In:             simpleCallRequest,
			ExpectedProps:  []interface{}{simpleCall, simpleContext},
			ExpectedString: "call: /some-path, context: bot_access_token: ***nXYZ, bot_user_id: id_of_bot_user, is_not_submit: true",
		},
		"full CallRequest": {
			In:             fullCallRequest,
			ExpectedProps:  []interface{}{fullCall, simpleContext, "values", "vkey1,vkey2"},
			ExpectedString: "call: /some-path, expand: acting_user_access_token:all,channel:summary,oauth2_app:all,user:all, state: key1,key2, context: bot_access_token: ***nXYZ, bot_user_id: id_of_bot_user, is_not_submit: true, values: vkey1,vkey2",
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
