// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsclient

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

const (
	// appsS3BucketEnvVarName determines an environment variable.
	// Variable saves address of apps S3 bucket name
	appsS3BucketEnvVarName = "MM_APPS_S3_BUCKET"

	// defaultBucketName is the default s3 bucket name used to store app data.
	defaultBucketName = "mattermost-apps-bucket"
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

func (c *client) CreateS3Bucket(bucket string) error {
	_, err := c.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create bucket")
	}

	return nil
}

// S3BucketExists return true if a bucket with the given name exists.
// Otherwise it returns false.
func (c *client) S3BucketExists(name string) (bool, error) {
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

func BucketWithDefaults(name string) string {
	if name != "" {
		return name
	}
	name = os.Getenv(appsS3BucketEnvVarName)
	if name != "" {
		return name
	}
	return defaultBucketName
}
