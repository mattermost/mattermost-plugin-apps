// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

const PrevVersion = "prev_version"

// SynchronizeInstalledApps synchronizes installed apps with known manifests,
// performing OnVersionChanged call on the App as needed.
func (p *Proxy) SynchronizeInstalledApps() error {
	installed := p.store.App.AsMap()
	println(fmt.Sprintf("installed = %v", installed))
	listed := p.store.Manifest.AsMap()
	println(fmt.Sprintf("listed = %v", listed))

	diff := map[apps.AppID]*apps.App{}
	for _, app := range installed {
		m := listed[app.AppID]

		// exclude unlisted apps, or those that need no action.
		if m == nil || app.Version == m.Version {
			continue
		}

		diff[app.AppID] = app
	}

	for _, app := range diff {
		m := listed[app.AppID]
		values := map[string]string{
			PrevVersion: string(app.Version),
		}

		// Store the new manifest to update the current mappings of the App
		app.Manifest = *m
		err := p.store.App.Save(app)
		if err != nil {
			return err
		}

		// Call OnVersionChanged the function of the app. It should be called only once
		if app.OnVersionChanged != nil {
			err := p.callOnce(func() error {
				creq := &apps.CallRequest{
					Call:   *app.OnVersionChanged,
					Values: map[string]interface{}{},
					Context: &apps.Context{
						AppID: app.AppID,
					},
				}
				for k, v := range values {
					creq.Values[k] = v
				}

				resp := p.Call("", creq)
				if resp.Type == apps.CallResponseTypeError {
					return errors.Wrapf(resp, "call %s failed", creq.Path)
				}
				return nil
			})
			if err != nil {
				p.mm.Log.Error("failed in callOnce:OnVersionChanged", "app_id", app.AppID, "err", err.Error())
			}
		}
	}

	return nil
}

func (p *Proxy) callOnce(f func() error) error {
	// Delete previous job
	if err := p.mm.KV.Delete(config.KeyCallOnce); err != nil {
		return errors.Wrap(err, "can't delete key")
	}
	// Ensure all instances run this
	time.Sleep(10 * time.Second)

	p.callOnceMutex.Lock()
	defer p.callOnceMutex.Unlock()
	value := 0
	if err := p.mm.KV.Get(config.KeyCallOnce, &value); err != nil {
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
	ok, err := p.mm.KV.Set(config.KeyCallOnce, value)
	if err != nil {
		return errors.Wrapf(err, "can't set key %s to %d", config.KeyCallOnce, value)
	}
	if !ok {
		return errors.Errorf("can't set key %s to %d", config.KeyCallOnce, value)
	}
	return nil
}
