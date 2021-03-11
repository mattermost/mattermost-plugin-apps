// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

const PrevVersion = "prev_version"

// SynchronizeInstalledApps synchronizes installed apps with known manifests,
// performing OnVersionChanged call on the App as needed.
func (adm *Admin) SynchronizeInstalledApps() error {
	installed := adm.GetInstalledApps()
	listed := adm.GetListedApps("")

	diff := map[apps.AppID]*apps.App{}
	for _, app := range installed {
		l := listed[app.AppID]

		// exclude unlisted apps, or those that need no action.
		if l == nil || app.Version == l.Manifest.Version {
			continue
		}

		diff[app.AppID] = app
	}

	// call onInstanceStartup. App migration happens here
	for _, app := range diff {
		l := listed[app.AppID]
		values := map[string]string{
			PrevVersion: string(app.Version),
		}

		// Store the new manifest to update the current mappings of the App
		app.Manifest = *l.Manifest
		err := adm.store.App().Save(app)
		if err != nil {
			return err
		}

		// Call OnVersionChanged the function of the app. It should be called only once
		if app.OnVersionChanged != nil {
			err := adm.callOnce(func() error {
				creq := &apps.CallRequest{
					Call:   *app.OnVersionChanged,
					Values: map[string]interface{}{},
				}
				for k, v := range values {
					creq.Values[k] = v
				}

				resp := adm.proxy.Call(adm.adminToken, creq)
				if resp.Type == apps.CallResponseTypeError {
					return errors.Wrapf(resp, "call %s failed", creq.Path)
				}
				return nil
			})
			if err != nil {
				adm.mm.Log.Error("failed in callOnce:OnVersionChanged", "app_id", app.AppID, "err", err.Error())
			}
		}
	}

	return nil
}

// func (adm *Admin) callOnStartupOnceWithValues(app *apps.App, values map[string]string) {
// 	// Call OnVersionChanged the function of the app. It should be called only once
// 	f := func() error {
// 		if err := adm.expandedCall(app, app.Manifest.OnVersionChanged, values); err != nil {
// 			adm.mm.Log.Error("Can't call OnVersionChanged func of the app", "app_id", app.Manifest.AppID, "err", err.Error())
// 		}
// 		return nil
// 	}
// 	if err := adm.callOnce(f); err != nil {
// 		adm.mm.Log.Error("Can't callOnce the OnVersionChanged func of the app", "app_id", app.Manifest.AppID, "err", err.Error())
// 	}
// }

func (adm *Admin) callOnce(f func() error) error {
	// Delete previous job
	if err := adm.mm.KV.Delete(api.KeyCallOnce); err != nil {
		return errors.Wrap(err, "can't delete key")
	}
	// Ensure all instances run this
	time.Sleep(10 * time.Second)

	adm.mutex.Lock()
	defer adm.mutex.Unlock()
	value := 0
	if err := adm.mm.KV.Get(api.KeyCallOnce, &value); err != nil {
		return err
	}
	if value != 0 {
		// job is already run by other instance
		return nil
	}

	// job is should be run by this instance
	if err := f(); err != nil {
		return errors.Wrap(err, "can't run the job")
	}
	value = 1
	ok, err := adm.mm.KV.Set(api.KeyCallOnce, value)
	if err != nil {
		return errors.Wrapf(err, "can't set key %s to %d", api.KeyCallOnce, value)
	}
	if !ok {
		return errors.Errorf("can't set key %s to %d", api.KeyCallOnce, value)
	}
	return nil
}
