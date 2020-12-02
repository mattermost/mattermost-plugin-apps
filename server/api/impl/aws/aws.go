// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	log "github.com/sirupsen/logrus"
)

// Service hold AWS clients for each service.
type Service struct {
	conf   api.Configurator
	logger log.FieldLogger
}

func NewAWS(conf api.Configurator) *Service {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	// Output to stdout instead of the default stderr.
	logger.SetOutput(os.Stdout)

	return &Service{
		conf:   conf,
		logger: logger,
	}
}

func (s *Service) newSession() *session.Session {
	// Make a copy of the cached session for logging
	sess := s.conf.GetConfig().AWSSession.Copy()
	sess.Handlers.Complete.PushFront(func(r *request.Request) {
		if r.HTTPResponse != nil && r.HTTPRequest != nil {
			var buffer bytes.Buffer
			l := s.logger.WithFields(log.Fields{
				"aws-service-id":     r.ClientInfo.ServiceID,
				"aws-operation-name": r.Operation.Name,
			})

			buffer.WriteString(fmt.Sprintf("[aws] %s %s (%s)", r.HTTPRequest.Method, r.HTTPRequest.URL.String(), r.HTTPResponse.Status))

			paramBytes, err := json.Marshal(r.Params)
			if err != nil {
				buffer.WriteString(err.Error())
			} else {
				l = l.WithField("params", string(paramBytes))
			}
			l.Debug(buffer.String())
		}
	})
	return sess
}

func (s *Service) lambda() *lambda.Lambda {
	return lambda.New(s.newSession(), aws.NewConfig().WithLogLevel(aws.LogDebugWithRequestErrors))
}

func (s *Service) iam() *iam.IAM {
	return iam.New(s.newSession())
}
