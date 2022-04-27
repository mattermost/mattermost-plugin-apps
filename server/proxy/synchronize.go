// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const PrevVersion = "prev_version"

// SynchronizeInstalledApps synchronizes installed apps with known manifests,
// performing OnVersionChanged call on the App as needed.
func (p *Proxy) SynchronizeInstalledApps() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	mm := p.conf.MattermostAPI()
	r := incoming.NewRequest(mm, p.conf, utils.NewPluginLogger(mm), p.sessionService, incoming.WithCtx(ctx))

	installed := p.store.App.AsMap()
	listed := p.store.Manifest.AsMap()

	diff := map[apps.AppID]apps.App{}
	for _, app := range installed {
		m, ok := listed[app.AppID]

		// exclude unlisted apps, or those that need no action.
		if !ok || app.Version == m.Version {
			continue
		}

		diff[app.AppID] = app
	}

	for id := range diff {
		r = r.Clone()
		r.SetAppID(id)

		app := diff[id]
		m := listed[app.AppID]

		// Store the new manifest to update the current mappings of the App
		app.Manifest = m
		err := p.store.App.Save(r, app)
		if err != nil {
			return err
		}

		// Call OnVersionChanged the function of the app. It should be called only once
		if app.OnVersionChanged != nil {
			err := p.callOnce(func() error {
				resp := p.call(r, app, *app.OnVersionChanged, nil, PrevVersion, app.Version)
				if resp.Type == apps.CallResponseTypeError {
					return errors.Wrapf(resp, "call %s failed", app.OnVersionChanged.Path)
				}
				return nil
			})
			if err != nil {
				r.Log.WithError(err).Errorw("failed in callOnce:OnVersionChanged",
					"app_id", app.AppID)
			}
		}
	}

	return nil
}

func (p *Proxy) callOnce(f func() error) error {
	mm := p.conf.MattermostAPI()
	// Delete previous job
	if err := mm.KV.Delete(config.KVCallOnceKey); err != nil {
		return errors.Wrap(err, "can't delete key")
	}
	// Ensure all instances run this
	time.Sleep(10 * time.Second)

	p.callOnceMutex.Lock()
	defer p.callOnceMutex.Unlock()
	value := 0
	if err := mm.KV.Get(config.KVCallOnceKey, &value); err != nil {
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
	ok, err := mm.KV.Set(config.KVCallOnceKey, value)
	if err != nil {
		return errors.Wrapf(err, "can't set key %s to %d", config.KVCallOnceKey, value)
	}
	if !ok {
		return errors.Errorf("can't set key %s to %d", config.KVCallOnceKey, value)
	}
	return nil
}
