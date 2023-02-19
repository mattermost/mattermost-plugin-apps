// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestUninstallApp(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		u                    uninstaller
		expectedError        string
		expectedMessage      string
		expectedForceError   string
		expectedForceMessage string
	}{
		{
			name:                 "happy path",
			u:                    uninstaller{},
			expectedMessage:      "Uninstalled appID (test app)",
			expectedForceMessage: "Uninstalled appID (test app)",
		},
		{
			name: "deleteApp error",
			u: uninstaller{
				deleteApp: func() error { return errors.New("deleteApp error") },
			},
			expectedError:      "disabled app after failing to clean up its data: failed to delete app: deleteApp error",
			expectedForceError: "disabled app after failing to clean up its data: failed to delete app: deleteApp error",
		},
		{
			name: "deleteAppData error",
			u: uninstaller{
				deleteAppData: func() error { return errors.New("deleteAppData error") },
			},
			expectedError:        "disabled app after failing to clean up its data: appID: failed to clean app's persisted data: deleteAppData error",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 1 error occurred:\n\t* appID: failed to clean app's persisted data: deleteAppData error\n\n",
		},
		{
			name: "deleteMattermostOAuth2App",
			u: uninstaller{
				deleteMattermostOAuth2App: func() error { return errors.New("deleteMattermostOAuth2App error") },
			},
			expectedError:        "disabled app after failing to clean up its data: appID: failed to delete Mattermost OAuth2 app record: deleteMattermostOAuth2App error",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 1 error occurred:\n\t* appID: failed to delete Mattermost OAuth2 app record: deleteMattermostOAuth2App error\n\n",
		},
		{
			name: "disableApp",
			u: uninstaller{
				// need to trigger disableApp, so return a cleanup error
				revokeSessionsForApp: func() error { return errors.New("revokeSessionsForApp error") },
				disableApp:           func() error { return errors.New("disableApp error") },
			},
			expectedError:        "2 errors occurred:\n\t* failed to clean up app data on uninstall: disabled app after failing to clean up its data: appID: failed to revoke sessions: revokeSessionsForApp error\n\t* failed to disable app after failing to clean up its data: disableApp error\n\n",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 1 error occurred:\n\t* appID: failed to revoke sessions: revokeSessionsForApp error\n\n",
		},
		{
			name: "disableBotAccount",
			u: uninstaller{
				disableBotAccount: func() error { return errors.New("disableBotAccount error") },
			},
			expectedError:        "disabled app after failing to clean up its data: appID: failed to disable Mattermost bot account: disableBotAccount error",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 1 error occurred:\n\t* appID: failed to disable Mattermost bot account: disableBotAccount error\n\n",
		},
		{
			name: "revokeSessionsForApp",
			u: uninstaller{
				revokeSessionsForApp: func() error { return errors.New("revokeSessionsForApp error") },
			},
			expectedError:        "disabled app after failing to clean up its data: appID: failed to revoke sessions: revokeSessionsForApp error",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 1 error occurred:\n\t* appID: failed to revoke sessions: revokeSessionsForApp error\n\n",
		},
		{
			name: "uninstallCall",
			u: uninstaller{
				uninstallCall: func() apps.CallResponse {
					return apps.CallResponse{
						Type: apps.CallResponseTypeError,
						Text: "uninstallCall error",
					}
				},
			},
			expectedError:        "appID: app canceled uninstall: uninstallCall error",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 1 error occurred:\n\t* appID: app canceled uninstall: uninstallCall error\n\n",
		},
		{
			name: "multiple errors",
			u: uninstaller{
				revokeSessionsForApp:      func() error { return errors.New("revokeSessionsForApp error") },
				deleteAppData:             func() error { return errors.New("deleteAppData error") },
				deleteMattermostOAuth2App: func() error { return errors.New("deleteMattermostOAuth2App error") },
			},
			expectedError:        "disabled app after failing to clean up its data: appID: failed to revoke sessions: revokeSessionsForApp error",
			expectedForceMessage: "Force-uninstalled appID (test app), despite error(s): 3 errors occurred:\n\t* appID: failed to revoke sessions: revokeSessionsForApp error\n\t* appID: failed to delete Mattermost OAuth2 app record: deleteMattermostOAuth2App error\n\t* appID: failed to clean app's persisted data: deleteAppData error\n\n",
		},
	} {
		for _, force := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s-%v", tc.name, force), func(t *testing.T) {
				tc.u.log = utils.NilLogger{}
				if tc.u.deleteApp == nil {
					tc.u.deleteApp = func() error { return nil }
				}
				if tc.u.deleteAppData == nil {
					tc.u.deleteAppData = func() error { return nil }
				}
				if tc.u.deleteMattermostOAuth2App == nil {
					tc.u.deleteMattermostOAuth2App = func() error { return nil }
				}
				if tc.u.disableApp == nil {
					tc.u.disableApp = func() error { return nil }
				}
				if tc.u.disableBotAccount == nil {
					tc.u.disableBotAccount = func() error { return nil }
				}
				if tc.u.revokeSessionsForApp == nil {
					tc.u.revokeSessionsForApp = func() error { return nil }
				}
				if tc.u.uninstallCall == nil {
					tc.u.uninstallCall = func() apps.CallResponse { return apps.CallResponse{} }
				}

				message, err := tc.u.uninstall(&apps.App{
					Manifest: apps.Manifest{
						AppID:       "appID",
						DisplayName: "test app",
						OnUninstall: apps.NewCall("/on_uninstall"),
					},
					MattermostOAuth2: &model.OAuthApp{},
				}, force)
				expectedErr := tc.expectedError
				expectedMessage := tc.expectedMessage
				if force {
					expectedErr = tc.expectedForceError
					expectedMessage = tc.expectedForceMessage
				}

				if expectedErr != "" {
					require.Error(t, err)
					require.Equal(t, expectedErr, err.Error())
				} else {
					require.NoError(t, err)
				}
				require.Equal(t, expectedMessage, message)
			})
		}
	}
}
