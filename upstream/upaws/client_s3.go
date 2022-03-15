// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
)

// ListS3 lists files in S3.
func (c *client) ListS3(bucket, prefix string) ([]string, error) {
	result, err := c.s3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list bucket")
	}
	keys := []string{}
	for _, o := range result.Contents {
		keys = append(keys, *o.Key)
	}
	return keys, nil
}

// GetS3 downloads files from S3.
func (c *client) GetS3(ctx context.Context, bucket, item string) ([]byte, error) {
	var buffer aws.WriteAtBuffer
	_, err := c.s3Down.DownloadWithContext(ctx, &buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file")
	}
	return buffer.Bytes(), nil
}

// UploadS3 uploads file to a specific S3 bucket
func (c *client) UploadS3(bucket, key string, body io.Reader, publicRead bool) (string, error) {
	_, err := c.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to upload file")
	}

	if publicRead {
		_, err = c.s3.PutObjectAcl(&s3.PutObjectAclInput{
			ACL:    aws.String("public-read"),
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return "", err
		}
	}

	u := url.URL{
		Scheme: `https`,
		Host:   fmt.Sprintf(`s3-%s.amazonaws.com`, c.region),
		Path:   fmt.Sprintf(`/%s/%s`, bucket, key),
	}
	return u.String(), nil
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

// ParseManifestS3Name parses the AppID and AppVersion out of an S3 key.
func ParseS3ManifestName(key string) (apps.AppID, apps.AppVersion, error) {
	i := strings.LastIndex(key, "_")
	if i == -1 || !strings.HasSuffix(key, ".json") {
		return "", "", errors.New("not a manifest file")
	}
	key = strings.TrimSuffix(key, ".json")
	id := key[:i]
	version := key[i+1:]
	return apps.AppID(id), apps.AppVersion(version), nil
}

// S3StaticName generates key for a specific asset in S3,
// key can be 1024 characters long.
func S3StaticName(appID apps.AppID, version apps.AppVersion, name string) string {
	sanitizedName := strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%s/%s_%s_app/%s", path.StaticFolder, appID, version, sanitizedName)
}

func S3BucketName() string {
	name := os.Getenv(S3BucketEnvVar)
	if name != "" {
		return name
	}
	return DefaultS3Bucket
}
