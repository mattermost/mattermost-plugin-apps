package restapi

import (
	"net/http"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"

	"github.com/mattermost/mattermost-server/v5/model"
)

type MarketplaceApp struct {
	Installed bool                     `json:"installed"`
	Labels    []model.MarketplaceLabel `json:"labels,omitempty"`
	Manifest  apps.Manifest            `json:"manifest"`
}

func (a *restapi) handleGetMarketplace(w http.ResponseWriter, req *http.Request, actingUserID string) {
	filter := req.URL.Query().Get("filter")

	apps, err := a.getApps()
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

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

func (a *restapi) getApps() ([]MarketplaceApp, error) {
	// registeredApps, _, err := a.api.Admin.ListApps()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "Failed to list local apps")
	// }

	// result := make([]MarketplaceApp, len(registeredApps))
	// for i := 0; i < len(registeredApps); i++ {
	// 	result[i] = MarketplaceApp{
	// 		Manifest:  *registeredApps[i].Manifest,
	// 		Installed: registeredApps[i].Status == apps.AppStatusInstalled,
	// 	}
	// }

	// return result, nil
	return nil, nil
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
