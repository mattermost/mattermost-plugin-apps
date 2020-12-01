// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	log "github.com/sirupsen/logrus"
)

// AWS is responsible for AWS management
type AWS interface {
	// releaseURL should contain a zip with lambda functions' zip files and a `manifest.json`
	// ~/my_app.zip
	//  |-- manifest.json
	//  |-- my_nodejs_function.zip
	//      |-- index.js
	//      |-- node-modules
	//          |-- async
	//          |-- aws-sdk
	//  |-- my_python_function.zip
	//      |-- lambda_function.py
	//      |-- __pycache__
	//      |-- certifi/
	InstallApp(releaseURL string) error
	InvokeFunction(functionName string, request interface{}) ([]byte, error)
}

// Client is a client for interacting with AWS resources.
type Client struct {
	logger  log.FieldLogger
	service *Service
	config  *aws.Config
	mux     *sync.Mutex
}

// Service hold AWS clients for each service.
type Service struct {
	lambda *lambda.Lambda
	iam    *iam.IAM
}

// NewAWSClientWithConfig returns a new instance of Client with a custom configuration.
func NewAWSClientWithConfig(config *aws.Config, logger log.FieldLogger) *Client {
	return &Client{
		logger: logger,
		config: config,
		mux:    &sync.Mutex{},
	}
}

// NewService creates a new instance of Service.
func NewService(sess *session.Session) *Service {
	return &Service{
		lambda: lambda.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithRequestErrors)),
		iam:    iam.New(sess),
	}
}

// Service contructs an AWS session if not yet successfully done and returns AWS clients.
func (c *Client) Service() *Service {
	if c.service == nil {
		sess, err := NewAWSSessionWithLogger(c.config, c.logger.WithField("aws", "client"))
		if err != nil {
			c.logger.WithError(err).Error("failed to initialize AWS session")
			// Calls to AWS will fail until a healthy session is acquired.
			return NewService(&session.Session{})
		}

		c.mux.Lock()
		c.service = NewService(sess)
		c.mux.Unlock()
	}

	return c.service
}

// NewAWSSessionWithLogger initializes an AWS session instance with logging handler for debuging only.
func NewAWSSessionWithLogger(config *aws.Config, logger log.FieldLogger) (*session.Session, error) {
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
				buffer.WriteString(err.Error())
			} else {
				logger = logger.WithField("params", string(paramBytes))
			}

			logger = logger.WithFields(log.Fields{
				"aws-service-id":     r.ClientInfo.ServiceID,
				"aws-operation-name": r.Operation.Name,
			})

			logger.Debug(buffer.String())
		}
	})

	return awsSession, nil
}
