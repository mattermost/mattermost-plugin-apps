// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
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

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type ARN string

func (arn ARN) AWSString() *string {
	return aws.String(string(arn))
}

type Name string

func (n Name) AWSString() *string {
	return aws.String(string(n))
}

// Client is an authenticated client for interacting with AWS resources. It
// provides a thin layer on top of aws-sdk-go, and contains all AWS
// dependencies.
type Client interface {
	// Proxy methods
	GetS3(bucket, item string) ([]byte, error)
	InvokeLambda(name string, invocationType string, payload []byte) ([]byte, error)

	// Admin methods
	AddResourcesToPolicyDocument(*iam.Policy, []ARN) (string, error)
	AddUserToGroup(u, g Name) error
	AttachGroupPolicy(g Name, p ARN) error
	AttachRolePolicy(roleName Name, policyARN ARN) error
	CreateAccessKey(user Name) (string, string, error)
	CreateGroup(name Name) (ARN, error)
	CreateLambda(zipFile io.Reader, function, handler, runtime string, role ARN) (ARN, error)
	CreateOrUpdateLambda(zipFile io.Reader, function, handler, runtime string, role ARN) (ARN, error)
	CreatePolicy(name Name, data string) (ARN, error)
	CreateRole(name Name) (ARN, error)
	CreateS3Bucket(bucket string) error
	CreateUser(name Name) (ARN, error)
	DeleteAccessKeys(user Name, accessKeyID string) error
	DeleteGroup(Name) error
	DeletePolicy(ARN) error
	DeleteRole(name Name) error
	DeleteS3Bucket(name string) error
	DeleteUser(name Name) error
	DetachGroupPolicy(g Name, p ARN) error
	ExistsS3Bucket(name string) (bool, error)
	FindGroup(name Name) (ARN, error)
	FindPolicy(policyName Name) (*iam.Policy, error)
	FindRole(name Name) (ARN, error)
	FindUser(name Name) (ARN, error)
	RemoveUserFromGroup(u, g Name) error
	UploadS3(bucket, key string, body io.Reader, publicRead bool) (string, error)
}

type client struct {
	lambda     lambdaiface.LambdaAPI
	iam        iamiface.IAMAPI
	s3Down     s3manageriface.DownloaderAPI
	s3Uploader s3manageriface.UploaderAPI
	s3         s3iface.S3API

	log    utils.Logger
	region string
}

func MakeClient(awsAccessKeyID, awsSecretAccessKey, region string, log utils.Logger) (Client, error) {
	awsConfig := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	}

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	awsSession.Handlers.Complete.PushFront(func(r *request.Request) {
		if r.HTTPResponse != nil && r.HTTPRequest != nil {
			log.Debugw("AWS request",
				"method", r.HTTPRequest.Method,
				"url", r.HTTPRequest.URL.String(),
				"status", r.HTTPResponse.Status,
				"aws-service-id", r.ClientInfo.ServiceID,
				"aws-operation-name", r.Operation.Name)
		}
	})

	c := &client{
		lambda:     lambda.New(awsSession, awsConfig),
		iam:        iam.New(awsSession),
		s3Down:     s3manager.NewDownloader(awsSession),
		s3Uploader: s3manager.NewUploader(awsSession),
		s3:         s3.New(awsSession),
		log:        log,
		region:     region,
	}

	return c, nil
}
