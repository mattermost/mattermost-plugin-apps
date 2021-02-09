package restapi

import (
	"net/http"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

type MarketplaceApp struct {
	Installed bool                     `json:"installed"`
	Labels    []model.MarketplaceLabel `json:"labels,omitempty"`
	Manifest  apps.Manifest            `json:"manifest"`
}

func (a *restapi) handleGetMarketplace(w http.ResponseWriter, req *http.Request, actingUserID string) {
	filter := req.URL.Query().Get("filter")

	localApps, err := a.getLocalApps()
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	remoteApps, err := a.getRemoteApps()
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	apps := mergeApps(remoteApps, localApps)

	// Filter plugins.
	var result []MarketplaceApp
	for _, a := range apps {
		if appMatchesFilter(a.Manifest, filter) {
			result = append(result, a)
		}
	}

	// Sort result alphabetically.
	sort.SliceStable(result, func(i, j int) bool {
		return strings.ToLower(result[i].Manifest.DisplayName) <
			strings.ToLower(result[j].Manifest.DisplayName)
	})

	httputils.WriteJSON(w, result)
}

func (a *restapi) getLocalApps() ([]MarketplaceApp, error) {
	localApps, _, err := a.api.Admin.ListApps()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list local apps")
	}

	result := make([]MarketplaceApp, len(localApps))
	for i := 0; i < len(localApps); i++ {
		result[i] = MarketplaceApp{
			Manifest:  *localApps[i].Manifest,
			Installed: localApps[i].Status == apps.AppStatusInstalled,
		}
	}

	return result, nil
}

func (a *restapi) getRemoteApps() ([]MarketplaceApp, error) {
	m := apps.Manifest{
		AppID:       "zendesk",
		Type:        apps.AppTypeHTTP,
		DisplayName: "Zendesk",
		Description: "A Zendesk App.",
		HomepageURL: "https://github.com/mattermost/mattermost-app-zendesk",
		HTTPRootURL: "http://localhost:4000/mattermost/manifest.json",
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsUser,
			apps.PermissionActAsBot,
		},
		RequestedLocations: apps.Locations{
			apps.LocationCommand,
			apps.LocationChannelHeader,
			apps.LocationInPost,
			apps.LocationPostMenu,
		},
	}

	result := []MarketplaceApp{{
		Manifest:  m,
		Installed: false,
	}}

	return result, nil
}

// mergeApps merges two slices of marketplace apps.
// If two items have the same id, the one from the first slice is keeped.
func mergeApps(a []MarketplaceApp, b []MarketplaceApp) []MarketplaceApp {
	appMap := map[string]*MarketplaceApp{}
	for i := range a {
		id := string(a[i].Manifest.AppID)
		appMap[id] = &a[i]
	}

	for i := range b {
		id := string(b[i].Manifest.AppID)
		if appMap[id] == nil {
			appMap[id] = &b[i]
		}
	}

	result := []MarketplaceApp{}
	for _, a := range appMap {
		result = append(result, *a)
	}

	return result
}

// Copied from Mattermost Server
func appMatchesFilter(manifest apps.Manifest, filter string) bool {
	filter = strings.TrimSpace(strings.ToLower(filter))

	if filter == "" {
		return true
	}

	if strings.ToLower(string(manifest.AppID)) == filter {
		return true
	}

	if strings.Contains(strings.ToLower(manifest.DisplayName), filter) {
		return true
	}

	if strings.Contains(strings.ToLower(manifest.Description), filter) {
		return true
	}

	return false
}
