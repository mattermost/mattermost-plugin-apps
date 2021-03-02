// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

const (
	// DefaultRegion describes default region in aws
	DefaultRegion = "us-east-2"

	// appsS3BucketEnvVarName determines an environment variable.
	// Variable saves address of apps S3 bucket name
	appsS3BucketEnvVarName = "MM_APPS_S3_BUCKET"

	// defaultBucketName is the default s3 bucket name used to store app data.
	defaultBucketName = "mattermost-apps-bucket"
)

// Client is a client for interacting with AWS resources.
type Client struct {
	logger       log
	service      *Service
	config       *aws.Config
	mux          *sync.Mutex
	AppsS3Bucket string
}

// Service hold AWS clients for each service.
type Service struct {
	lambda       lambdaiface.LambdaAPI
	iam          iamiface.IAMAPI
	s3           s3iface.S3API
	s3Downloader s3manageriface.DownloaderAPI
	s3Uploader   s3manageriface.UploaderAPI
}

type log interface {
	Error(message string, keyValuePairs ...interface{})
	Warn(message string, keyValuePairs ...interface{})
	Info(message string, keyValuePairs ...interface{})
	Debug(message string, keyValuePairs ...interface{})
}

// NewAWSClientWithConfig returns a new instance of Client with a custom configuration.
func NewAWSClientWithConfig(config *aws.Config, bucket string, logger log) *Client {
	return &Client{
		logger:       logger,
		config:       config,
		mux:          &sync.Mutex{},
		AppsS3Bucket: bucket,
	}
}

func NewAWSClient(awsAccessKeyID, awsSecretAccessKey string, logger log) *Client {
	config := createAWSConfig(awsAccessKeyID, awsSecretAccessKey)
	bucket := os.Getenv(appsS3BucketEnvVarName)
	if bucket == "" {
		bucket = defaultBucketName
	}
	return NewAWSClientWithConfig(config, bucket, logger)
}

func createAWSConfig(awsAccessKeyID, awsSecretAccessKey string) *aws.Config {
	creds := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")

	return &aws.Config{
		Region:      aws.String(DefaultRegion),
		Credentials: creds,
	}
}

// NewService creates a new instance of Service.
func NewService(sess *session.Session) *Service {
	return &Service{
		lambda:       lambda.New(sess, aws.NewConfig()),
		iam:          iam.New(sess),
		s3:           s3.New(sess),
		s3Downloader: s3manager.NewDownloader(sess),
		s3Uploader:   s3manager.NewUploader(sess),
	}
}

// Service contructs an AWS session if not yet successfully done and returns AWS clients.
func (c *Client) Service() *Service {
	if c.service == nil {
		c.newAWSSession()
	}

	return c.service
}

// RefreshService refreshes aws session using with new access key and secret.
func (c *Client) RefreshService(awsAccessKeyID, awsSecretAccessKey string) {
	config := createAWSConfig(awsAccessKeyID, awsSecretAccessKey)
	c.refreshService(config)
}

func (c *Client) refreshService(newConfig *aws.Config) {
	c.config = newConfig
	c.newAWSSession()
}

func (c *Client) newAWSSession() {
	sess, err := NewAWSSessionWithLogger(c.config, c.logger)
	if err != nil {
		c.logger.Error("failed to initialize AWS session", "err", err.Error())
		// Calls to AWS will fail until a healthy session is acquired.
		c.mux.Lock()
		c.service = NewService(&session.Session{})
		c.mux.Unlock()
	} else {
		c.mux.Lock()
		c.service = NewService(sess)
		c.mux.Unlock()
	}
}

// NewAWSSessionWithLogger initializes an AWS session instance with logging handler for debuging only.
func NewAWSSessionWithLogger(config *aws.Config, logger log) (*session.Session, error) {
	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	awsSession.Handlers.Complete.PushFront(func(r *request.Request) {
		if r.HTTPResponse != nil && r.HTTPRequest != nil {
			var buffer bytes.Buffer

			buffer.WriteString(fmt.Sprintf("[aws] %s %s (%s)", r.HTTPRequest.Method, r.HTTPRequest.URL.String(), r.HTTPResponse.Status))

			paramBytes, err := json.Marshal(r.Params)
			if err != nil {
				buffer.WriteString(fmt.Sprintf("error: %s", err.Error()))
			} else {
				buffer.WriteString(fmt.Sprintf("params: %s", string(paramBytes)))
			}

			buffer.WriteString(fmt.Sprintf("aws-service-id: %s. aws-operation-name: %s", r.ClientInfo.ServiceID, r.Operation.Name))
			logger.Debug(buffer.String())
		}
	})

	return awsSession, nil
}
