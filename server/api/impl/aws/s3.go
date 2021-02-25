// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// S3FileDownload is used to download files from the S3.
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

// S3FileUpload uploads file to a specific S3 bucket
func (c *Client) S3FileUpload(key string, body io.Reader) error {
	if _, err := c.service.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.AppsS3Bucket),
		Key:    aws.String(key),
		Body:   body,
	}); err != nil {
		return errors.Wrap(err, "failed to upload file")
	}
	return nil
}

// S3AssetDownload is used to download app's static assets from the S3.
func (c *Client) S3AssetDownload(item string) ([]byte, error) {
	return c.S3FileDownload(c.AppsS3Bucket, item)
}

// CreateBucket creates a new s3 bucket.
func (c *Client) CreateBucket(bucket string) error {
	_, err := c.Service().s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create bucket")
	}

	return nil
}

// CheckIfBucketExists return true if a bucket with the given name exists.
// Otherwise it returns false.
func (c *Client) CheckIfBucketExists(name string) (bool, error) {
	buckets, err := c.Service().s3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return false, errors.Wrap(err, "failed to list buckets")
	}

	for _, b := range buckets.Buckets {
		if *b.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetManifest returns a manifest file for an app from the S3.
func (c *Client) GetManifest(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error) {
	key := getManifestFileKey(appID, version)
	data, err := c.S3FileDownload(c.AppsS3Bucket, key)
	if err != nil {
		return nil, errors.Wrapf(err, "can't download manifest %s/%s", c.AppsS3Bucket, key)
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

// SaveManifest saves manifest file in S3
func (c *Client) SaveManifest(manifest *apps.Manifest) error {
	data, err := json.Marshal(manifest)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", manifest.AppID)
	}
	buffer := bytes.NewBuffer(data)

	key := getManifestFileKey(manifest.AppID, manifest.Version)
	if err := c.S3FileUpload(key, buffer); err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", manifest.AppID)
	}
	return nil
}
