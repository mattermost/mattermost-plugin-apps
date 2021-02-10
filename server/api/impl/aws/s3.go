// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// S3FileDownload is used to download files from the S3
func (c *Client) S3FileDownload(bucket, item string) ([]byte, error) {
	var buffer aws.WriteAtBuffer
	_, err := c.Service().s3Downloader.Download(&buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file")
	}
	return buffer.Bytes(), nil
}

// GetManifest returns a manifest file for an app from the S3
func (c *Client) GetManifest(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error) {
	manifestFileName := fmt.Sprintf("manifest_%s_%s", appID, version)
	data, err := c.S3FileDownload(c.appsS3Bucket, manifestFileName)
	if err != nil {
		return nil, errors.Wrapf(err, "can't download manifest %s/%s", c.appsS3Bucket, manifestFileName)
	}
	var manifest *apps.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if manifest == nil {
		return nil, errors.Errorf("manifest is nil for app %s", appID)
	}
	if manifest.AppID != appID {
		return nil, errors.Errorf("missmatched app ids while getting manifest %s != %s", manifest.AppID, appID)
	}
	return manifest, nil
}
