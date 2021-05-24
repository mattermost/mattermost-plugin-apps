// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// GetS3 downloads files from S3.
func (c *client) GetS3(bucket, item string) ([]byte, error) {
	var buffer aws.WriteAtBuffer
	_, err := c.s3Down.Download(&buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file")
	}
	return buffer.Bytes(), nil
}

// UploadS3 uploads file to a specific S3 bucket
func (c *client) UploadS3(bucket, key string, body io.Reader) error {
	if _, err := c.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}); err != nil {
		return errors.Wrap(err, "failed to upload file")
	}
	return nil
}

func (c *client) CreateS3Bucket(bucket string) error {
	_, err := c.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create bucket")
	}

	return nil
}

func (c *client) DeleteS3Bucket(bucket string) error {
	_, err := c.s3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete bucket")
	}
	return nil
}

// ExistsS3Bucket return true if a bucket with the given name exists.
// Otherwise it returns false.
func (c *client) ExistsS3Bucket(name string) (bool, error) {
	buckets, err := c.s3.ListBuckets(&s3.ListBucketsInput{})
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

// ManifestS3Name generates key for a specific manifest in S3,
// key can be 1024 characters long.
func S3ManifestName(appID apps.AppID, version apps.AppVersion) string {
	return fmt.Sprintf("manifests/%s_%s.json", appID, version)
}

// GenerateAssetS3Name generates key for a specific asset in S3,
// key can be 1024 characters long.
func S3StaticName(appID apps.AppID, version apps.AppVersion, name string) string {
	sanitizedName := strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%s/%s_%s_app/%s", apps.StaticFolder, appID, version, sanitizedName)
}

func S3BucketName() string {
	name := os.Getenv(S3BucketEnvVar)
	if name != "" {
		return name
	}
	return DefaultS3Bucket
}
