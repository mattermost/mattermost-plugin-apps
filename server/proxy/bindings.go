package proxy

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/logger"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/pkg/errors"
)

const (
	CacheKeyValueMaxRunes       = 50
	BindingsCacheKeyAllUsers    = "ALL_USERS"
	BindingsCacheKeyAllChannels = "ALL_CHANNELS"
	BindingsCachePrefix         = "bindings"
)

func mergeBindings(bb1, bb2 []*apps.Binding) []*apps.Binding {
	out := append([]*apps.Binding(nil), bb1...)

	for _, b2 := range bb2 {
		found := false
		for i, o := range out {
			if b2.AppID == o.AppID && b2.Location == o.Location {
				found = true

				// b2 overrides b1, if b1 and b2 have Bindings, they are merged
				merged := b2
				if len(o.Bindings) != 0 && b2.Call == nil {
					merged.Bindings = mergeBindings(o.Bindings, b2.Bindings)
				}
				out[i] = merged
			}
		}
		if !found {
			out = append(out, b2)
		}
	}
	return out
}

// GetBindings fetches bindings for all apps.
// We should avoid unnecessary logging here as this route is called very often.
func (p *Proxy) GetBindings(sessionID, actingUserID string, cc *apps.Context) ([]*apps.Binding, error) {
	allApps := store.SortApps(p.store.App.AsMap())
	all := make([][]*apps.Binding, len(allApps))

	var wg sync.WaitGroup
	for i, app := range allApps {
		wg.Add(1)
		go func(app *apps.App, i int) {
			defer wg.Done()
			all[i] = p.GetBindingsForApp(sessionID, actingUserID, cc, app)
		}(app, i)
	}
	wg.Wait()

	ret := []*apps.Binding{}
	for _, b := range all {
		ret = mergeBindings(ret, b)
	}

	return ret, nil
}

// GetBindingsForApp fetches bindings for a specific apps.
// We should avoid unnecessary logging here as this route is called very often.
func (p *Proxy) GetBindingsForApp(sessionID, actingUserID string, cc *apps.Context, app *apps.App) []*apps.Binding {
	if !p.AppIsEnabled(app) {
		return nil
	}

	logger := logger.New(&p.mm.Log).With(logger.LogContext{
		"app_id": cc.AppID,
	})

	appID := app.AppID
	appCC := *cc
	appCC.AppID = appID
	appCC.BotAccessToken = app.BotAccessToken

	var err error
	var bindings = []*apps.Binding{}
	bindings, err = p.CacheGetAllBindings(cc, appID)
	if err != nil || len(bindings) == 0 {
		bindingsCall := apps.DefaultBindings.WithOverrides(app.Bindings)
		bindingsRequest := &apps.CallRequest{
			Call:    *bindingsCall,
			Context: &appCC,
		}

		resp := p.Call(sessionID, actingUserID, bindingsRequest)
		if resp == nil || (resp.Type != apps.CallResponseTypeError && resp.Type != apps.CallResponseTypeOK) {
			logger.Debugf("Bindings response is nil or unexpected type.")
			return nil
		}

		if resp.Type == apps.CallResponseTypeError {
			logger.Debugf("Error getting bindings. Error: " + resp.Error())
			return nil
		}

		b, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(b, &bindings)
		if err != nil {
			logger.Debugf("Bindings are not of the right type.")
			return nil
		} else if storeErr := p.CacheSetBindings(cc, appID, bindings); storeErr != nil { // store the bindings to the cache
			p.mm.Log.Error(fmt.Sprintf("failed to store bindings to cache for %s: %v", appID, storeErr))
		}

		bindings = p.scanAppBindings(app, bindings, "")
	}

	return bindings
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (p *Proxy) scanAppBindings(app *apps.App, bindings []*apps.Binding, locPrefix apps.Location) []*apps.Binding {
	out := []*apps.Binding{}
	locationsUsed := map[apps.Location]bool{}
	labelsUsed := map[string]bool{}

	for _, appB := range bindings {
		// clone just in case
		b := *appB
		if b.Location == "" {
			b.Location = apps.Location(app.Manifest.AppID)
		}

		fql := locPrefix.Make(b.Location)
		allowed := false
		for _, grantedLoc := range app.GrantedLocations {
			if fql.In(grantedLoc) || grantedLoc.In(fql) {
				allowed = true
				break
			}
		}
		if !allowed {
			// p.mm.Log.Debug(fmt.Sprintf("location %s is not granted to app %s", fql, app.Manifest.AppID))
			continue
		}

		if fql.IsTop() {
			if locationsUsed[appB.Location] {
				continue
			}
			locationsUsed[appB.Location] = true
		} else {
			if b.Location == "" || b.Label == "" {
				continue
			}
			if locationsUsed[appB.Location] || labelsUsed[appB.Label] {
				continue
			}

			locationsUsed[appB.Location] = true
			labelsUsed[appB.Label] = true
			b.AppID = app.Manifest.AppID
		}

		if len(b.Bindings) != 0 {
			scanned := p.scanAppBindings(app, b.Bindings, fql)
			if len(scanned) == 0 {
				// We do not add bindings without any valid sub-bindings
				continue
			}
			b.Bindings = scanned
		}

		out = append(out, &b)
	}

	return out
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	p.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
}

func (p *Proxy) CacheSetBindings(cc *apps.Context, appID apps.AppID, bindings []*apps.Binding) error {
	groupedBindingsMap := map[string][][]byte{}

	for _, binding := range bindings {
		userID := BindingsCacheKeyAllUsers
		if binding.DependsOnUser && cc.ActingUserID != "" {
			userID = cc.ActingUserID
		}

		channelID := BindingsCacheKeyAllChannels
		if binding.DependsOnChannel && cc.ChannelID != "" {
			channelID = cc.ChannelID
		}

		valueBytes, err := json.Marshal(binding)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal value")
		}

		key := p.CacheBuildKey(BindingsCachePrefix, userID, channelID)
		bindingsForKey := groupedBindingsMap[key]
		bindingsForKey = append(bindingsForKey, valueBytes)
		groupedBindingsMap[key] = bindingsForKey
	}

	if storeErr := p.mm.AppsCache.Set(string(appID), groupedBindingsMap); storeErr != nil {
		p.mm.Log.Error(fmt.Sprintf("failed to store bindings to cache for %s: %v", appID, storeErr))
		return storeErr
	}

	return nil
}

func (p *Proxy) CacheGetAllBindings(cc *apps.Context, appID apps.AppID) ([]*apps.Binding, error) {
	bindings := []*apps.Binding{}

	keys := p.CacheBuildBindingsKeys(BindingsCachePrefix, cc.ActingUserID, cc.ChannelID)
	for _, key := range keys {
		tbindings, err := p.CacheGetBinding(cc, appID, key)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, tbindings...)
	}

	return bindings, nil
}

func (p *Proxy) CacheGetBinding(cc *apps.Context, appID apps.AppID, key string) ([]*apps.Binding, error) {
	bindings := []*apps.Binding{}

	var retErr error
	if outBindings, err := p.mm.AppsCache.Get(string(appID), key); err == nil {
		b := apps.Binding{}
		for _, outBinding := range outBindings {
			if err := json.Unmarshal(outBinding, &b); err != nil {
				p.mm.Log.Error(fmt.Sprintf("failed to unmarshal value for key %s", key))
				retErr = err
				break
			}
			bindings = append(bindings, &b)
		}
	}

	return bindings, retErr
}

func (p *Proxy) CacheDelete(appID apps.AppID, key string) error {
	return p.mm.AppsCache.Delete(string(appID), key)
}

func (p *Proxy) CacheEmpty(appID apps.AppID) error {
	return p.mm.AppsCache.DeleteAll(string(appID))
}

func (p *Proxy) CacheEmptyApps() []error {
	errors := []error{}

	allApps := store.SortApps(p.store.App.AsMap())
	for _, app := range allApps {
		if err := p.CacheEmpty(app.Manifest.AppID); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (p *Proxy) CacheBuildBindingsKeys(prefix string, userID string, channelID string) []string {
	keys := []string{}

	keys = append(keys, p.CacheBuildKey(prefix, BindingsCacheKeyAllUsers, BindingsCacheKeyAllChannels))

	if userID != "" {
		keys = append(keys, p.CacheBuildKey(prefix, userID, BindingsCacheKeyAllChannels))
	}

	if channelID != "" {
		keys = append(keys, p.CacheBuildKey(prefix, BindingsCacheKeyAllUsers, channelID))
	}

	if userID != "" && channelID != "" {
		keys = append(keys, p.CacheBuildKey(prefix, userID, channelID))
	}
	return keys
}

func (p *Proxy) CacheBuildKey(keyParts ...string) string {
	keyPartsHash := md5.Sum([]byte(strings.Join(keyParts, ":"))) // nolint:gosec
	key := base64.RawURLEncoding.EncodeToString(keyPartsHash[:])

	if utf8.RuneCountInString(key) > CacheKeyValueMaxRunes {
		return key[:CacheKeyValueMaxRunes]
	}

	return key
}

func (p *Proxy) CacheInvalidateBindings(cc *apps.Context, appID apps.AppID) error {
	userID := cc.ActingUserID
	channelID := cc.ChannelID

	if cc.ActingUserID == "" {
		userID = BindingsCacheKeyAllUsers
	}

	if cc.ChannelID == "" {
		channelID = BindingsCacheKeyAllChannels
	}

	key := p.CacheBuildKey(BindingsCachePrefix, userID, channelID)
	return p.CacheDelete(appID, key)
}
