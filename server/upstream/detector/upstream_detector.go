package detector

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream/upawslambda"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream/uphttp"
)

type Detector interface {
	UpstreamForApp(app *apps.App) (upstream.Upstream, error)
	AddBuiltinUpstream(appID apps.AppID, up upstream.Upstream)
}

type UpstreamDetector struct {
	builtin       map[apps.AppID]upstream.Upstream
	client        aws.Client
	s3AssetBucket string
}

func NewDetector(client aws.Client, s3AssetBucket string) *UpstreamDetector {
	return &UpstreamDetector{
		builtin:       map[apps.AppID]upstream.Upstream{},
		client:        client,
		s3AssetBucket: s3AssetBucket,
	}
}

func (d *UpstreamDetector) AddBuiltinUpstream(appID apps.AppID, up upstream.Upstream) {
	d.builtin[appID] = up
}

func (d *UpstreamDetector) UpstreamForApp(app *apps.App) (upstream.Upstream, error) {
	switch app.AppType {
	case apps.AppTypeHTTP:
		return uphttp.NewUpstream(app), nil

	case apps.AppTypeAWSLambda:
		return upawslambda.NewUpstream(app, d.client, d.s3AssetBucket), nil

	case apps.AppTypeBuiltin:
		up := d.builtin[app.AppID]
		if up == nil {
			return nil, errors.Errorf("builtin app not found: %s", app.AppID)
		}
		return up, nil

	default:
		return nil, errors.Errorf("not a valid app type: %s", app.AppType)
	}
}
