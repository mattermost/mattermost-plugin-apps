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

// Client is an authenticated client for interacting with AWS resources. It
// provides a thin layer on top of aws-sdk-go, and contains all AWS
// dependencies.
type Client interface {
	// Proxy methods
	GetS3(bucket, item string) ([]byte, error)
	InvokeLambda(name string, invocationType string, payload []byte) ([]byte, error)

	// Admin methods
	AddUserToGroup(u, g string) error
	AttachGroupPolicy(g, p string) error
	CreateAccessKey(u string) (string, string, error)
	CreateGroup(name string) (string, error)
	CreateLambda(zipFile io.Reader, function, handler, runtime, resource string) error
	CreateOrUpdateLambda(zipFile io.Reader, function, handler, runtime, resource string) error
	CreatePolicy(name, data string) (arn string, _ error)
	CreateS3Bucket(bucket string) error
	CreateUser(name string) (string, error)
	DeleteAccessKeys(u string) error
	DeleteGroup(name string) error
	DeletePolicy(name string) error
	DeleteS3Bucket(name string) error
	DeleteUser(name string) error
	DetachGroupPolicy(g, p string) error
	ExistsS3Bucket(name string) (bool, error)
	FindGroup(name string) (string, error)
	FindPolicy(policyName string) (*iam.Policy, error)
	FindUser(name string) (string, error)
	RemoveUserFromGroup(u, g string) error
	UploadS3(bucket, key string, body io.Reader) error
}

type client struct {
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

func MakeClient(awsAccessKeyID, awsSecretAccessKey, region string, logger Logger) (Client, error) {
	awsConfig := &aws.Config{
		Region:      aws.String(region),
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

				buffer.WriteString(fmt.Sprintf("[aws] %s %s (%s), ", r.HTTPRequest.Method, r.HTTPRequest.URL.String(), r.HTTPResponse.Status))
				buffer.WriteString(fmt.Sprintf("aws-service-id: %s, aws-operation-name: %s, ", r.ClientInfo.ServiceID, r.Operation.Name))

				paramBytes, err := json.Marshal(r.Params)
				if err != nil {
					buffer.WriteString(fmt.Sprintf("error: %s ", err.Error()))
				} else {
					pstr := string(paramBytes)
					if len(pstr) > 1024 {
						pstr = pstr[:1024] + "..."
					}
					buffer.WriteString(fmt.Sprintf("params: %s", pstr))
				}

				logger.Debug(buffer.String())
			}
		})
	}

	c := &client{
		lambda:     lambda.New(awsSession, awsConfig),
		iam:        iam.New(awsSession),
		s3Down:     s3manager.NewDownloader(awsSession),
		s3Uploader: s3manager.NewUploader(awsSession),
		s3:         s3.New(awsSession),
		logger:     logger,
	}

	return c, nil
}
