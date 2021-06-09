package upaws

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

type StaticUpstream struct {
	manifest  *apps.Manifest
	awsClient Client
	bucket    string
}

var _ upstream.StaticUpstream = (*StaticUpstream)(nil)

func NewStaticUpstream(m *apps.Manifest, awsClient Client, bucket string) *StaticUpstream {
	return &StaticUpstream{
		manifest:  m,
		awsClient: awsClient,
		bucket:    bucket,
	}
}

func (u *StaticUpstream) GetStatic(path string) (io.ReadCloser, int, error) {
	key := S3StaticName(u.manifest.AppID, u.manifest.Version, path)
	data, err := u.awsClient.GetS3(u.bucket, key)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "can't download from S3:bucket:%s, path:%s", u.bucket, path)
	}
	return io.NopCloser(bytes.NewReader(data)), http.StatusOK, nil
}
