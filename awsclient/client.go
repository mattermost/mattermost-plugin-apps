// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

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

// DefaultRegion describes default region in aws
const DefaultRegion = "us-east-2"

// Client is an authenticated client for interacting with AWS resources.
type Client interface {
	// Proxy methods

	GetS3(bucket, item string) ([]byte, error)
	InvokeLambda(name string, invocationType string, request interface{}) ([]byte, error)

	// Admin methods

	CreateLambda(zipFile io.Reader, function, handler, runtime, resource string) error
	CreateOrUpdateLambda(zipFile io.Reader, function, handler, runtime, resource string) error
	MakeLambdaFunctionDefaultPolicy() (string, error)

	CreateS3Bucket(bucket string) error
	UploadS3(bucket, key string, body io.Reader) error
	ExistsS3Bucket(name string) (bool, error)
}

type client struct {
	accessKeyID     string
	secretAccessKey string

	lambda     lambdaiface.LambdaAPI
	iam        iamiface.IAMAPI
	s3Down     s3manageriface.DownloaderAPI
	s3Uploader s3manageriface.UploaderAPI
	s3         s3iface.S3API

	logger Logger
}

type Logger interface {
	Error(message string, keyValuePairs ...interface{})
	Warn(message string, keyValuePairs ...interface{})
	Info(message string, keyValuePairs ...interface{})
	Debug(message string, keyValuePairs ...interface{})
}

func MakeClient(awsAccessKeyID, awsSecretAccessKey string, logger Logger) (Client, error) {
	awsConfig := &aws.Config{
		Region:      aws.String(DefaultRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	}

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	if logger != nil {
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
	}

	c := &client{
		accessKeyID:     awsAccessKeyID,
		secretAccessKey: awsSecretAccessKey,
		lambda:          lambda.New(awsSession, aws.NewConfig().WithLogLevel(aws.LogDebugWithRequestErrors)),
		iam:             iam.New(awsSession),
		s3Down:          s3manager.NewDownloader(awsSession),
		s3Uploader:      s3manager.NewUploader(awsSession),
		s3:              s3.New(awsSession),
		logger:          logger,
	}

	return c, nil
}
