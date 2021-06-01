// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"bytes"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type InitParams struct {
	Bucket                string
	Policy                Name
	User                  Name
	Group                 Name
	ExecuteRole           Name
	ShouldCreate          bool
	ShouldCreateAccessKey bool
}

type InitResult struct {
	Bucket          string
	PolicyARN       ARN
	UserARN         ARN
	GroupARN        ARN
	ExecuteRoleARN  ARN
	AccessKeyID     string
	AccessKeySecret string
}

func InitializeAWS(asAdmin Client, log Logger, params InitParams) (r *InitResult, err error) {
	r = &InitResult{}
	exists, err := asAdmin.ExistsS3Bucket(params.Bucket)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket %s must exist, please create one using cli or the console", params.Bucket)
	}
	if !exists {
		return nil, errors.Errorf("S3 bucket %q is not configured", params.Bucket)
	}
	log.Info("Using existing S3 bucket", "name", params.Bucket)
	r.Bucket = params.Bucket

	ensure := func(typ string, name Name, find, create func(Name) (ARN, error)) (ARN, error) {
		var arn ARN
		arn, err = find(name)
		if err != nil {
			if !errors.Is(err, utils.ErrNotFound) || !params.ShouldCreate {
				return "", errors.Wrapf(err, "failed to find %s %s", typ, name)
			}

			arn, err = create(name)
			if err != nil {
				return "", errors.Wrapf(err, "failed to create %s %s", typ, name)
			}
			log.Info("Created "+typ, "ARN", arn)
		} else {
			log.Info("Using existing "+typ, "ARN", arn)
		}
		return arn, nil
	}

	// Ensure user and group
	r.UserARN, err = ensure("user", params.User, asAdmin.FindUser, asAdmin.CreateUser)
	if err != nil {
		return nil, err
	}
	if params.ShouldCreateAccessKey {
		r.AccessKeyID, r.AccessKeySecret, err = asAdmin.CreateAccessKey(params.User)
		if err != nil {
			return nil, err
		}
		log.Info("Created access key")
	}
	r.GroupARN, err = ensure("group", params.Group, asAdmin.FindGroup, asAdmin.CreateGroup)
	if err != nil {
		return nil, err
	}

	// Ensure invoke policy that provides user access to the App's resources.
	r.PolicyARN, err = ensure("policy", params.Policy,
		func(name Name) (ARN, error) {
			var p *iam.Policy
			p, err = asAdmin.FindPolicy(name)
			if err != nil {
				return "", err
			}
			return ARN(*p.Arn), nil
		},
		func(name Name) (ARN, error) {
			out := &bytes.Buffer{}
			err = InvokePolicyDocumentTemplate.Execute(out, params)
			return asAdmin.CreatePolicy(name, out.String())
		})
	if err != nil {
		return nil, err
	}

	// Connect user-group-policy
	err = asAdmin.AttachGroupPolicy(params.Group, r.PolicyARN)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to attach %s to %s", params.Policy, params.Group)
	}
	log.Info("Attached policy to group", "policyARN", r.PolicyARN, "groupName", params.Group)
	err = asAdmin.AddUserToGroup(params.User, params.Group)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add user %s to %s", params.User, params.Group)
	}
	log.Info("Added user to group", "userName", params.User, "groupName", params.Group)

	// Create an execution role for the Apps' Lambdas. It uses
	// AWSLambdaBasicExecutionRole service execution policy.
	r.ExecuteRoleARN, err = ensure("execute role", params.ExecuteRole, asAdmin.FindRole, asAdmin.CreateRole)
	if err != nil {
		return nil, err
	}
	err = asAdmin.AttachRolePolicy(params.ExecuteRole, LambdaExecutionPolicyARN)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to attach %s to %s", LambdaExecutionPolicyARN, params.ExecuteRole)
	}
	log.Info("Attached AWSLambdaBasicExecutionRole policy to role", "roleName", params.ExecuteRole)

	return r, nil
}
