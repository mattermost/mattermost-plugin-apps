// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
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
