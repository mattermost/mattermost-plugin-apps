// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
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

func InitApps(asAdmin Client, log Logger, params InitParams) (r *InitResult, err error) {
	r = &InitResult{}
	exists, err := asAdmin.ExistsS3Bucket(params.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("S3 bucket %q is not configured", params.Bucket)
	}
	log.Info("using existing S3 bucket", "name", params.Bucket)
	r.Bucket = params.Bucket

	ensure := func(typ string, name Name, find, create func(Name) (ARN, error)) (ARN, error) {
		var arn ARN
		arn, err = find(name)
		if err != nil {
			if errors.Cause(err) != utils.ErrNotFound || !params.ShouldCreate {
				return "", err
			}
			arn, err = create(name)
			if err != nil {
				return "", err
			}
			log.Info("created "+typ, "ARN", arn)
		} else {
			log.Info("using existing "+typ, "ARN", arn)
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
		log.Info("created access key")
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
			return asAdmin.CreatePolicy(name, `{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Sid": "AllowLamda",
						"Effect": "Allow",
						"Action": [
							"lambda:InvokeFunction"
						],
						"Resource": [
							"arn:aws:lambda:::function:ping"
						]
					}
				]
			}`)
		})
	if err != nil {
		return nil, err
	}

	// Connect user-group-policy
	err = asAdmin.AttachGroupPolicy(params.Group, r.PolicyARN)
	if err != nil {
		return nil, err
	}
	log.Info("attached policy to group", "policyARN", r.PolicyARN, "groupName", params.Group)
	err = asAdmin.AddUserToGroup(params.User, params.Group)
	if err != nil {
		return nil, err
	}
	log.Info("added user to group", "userName", params.User, "groupName", params.Group)

	// Create an execution role for the Apps' Lambdas. It uses
	// AWSLambdaBasicExecutionRole service execution policy.
	r.ExecuteRoleARN, err = ensure("execute role", params.ExecuteRole, asAdmin.FindRole, asAdmin.CreateRole)
	if err != nil {
		return nil, err
	}
	err = asAdmin.AttachRolePolicy(params.ExecuteRole, LambdaExecutionPolicyARN)
	if err != nil {
		return nil, err
	}
	log.Info("attached AWSLambdaBasicExecutionRole policy to role", "roleName", params.ExecuteRole)

	return r, nil
}

func CleanApps(asAdmin Client, accessKeyID string, log Logger) error {
	delete := func(typ string, name Name, del func(Name) error) error {
		err := del(name)
		if err != nil {
			if errors.Cause(err) != utils.ErrNotFound {
				return err
			}
			log.Info("not found "+typ, "key", name)
		} else {
			log.Info("deleted "+typ, "key", name)
		}
		return nil
	}

	var err error
	err = asAdmin.RemoveUserFromGroup(DefaultUserName, DefaultGroupName)
	switch {
	case err == nil:
		log.Info("removed user from group", "user", DefaultUserName, "group", DefaultGroupName)
	case errors.Cause(err) == utils.ErrNotFound:
		// nothing to do
	default:
		return err
	}

	policy, err := asAdmin.FindPolicy(DefaultPolicyName)
	if err == nil {
		err = asAdmin.DetachGroupPolicy(DefaultGroupName, ARN(*policy.Arn))
		switch {
		case err == nil:
			log.Info("detached policy from group", "policy", DefaultPolicyName, "group", DefaultGroupName)
		case errors.Cause(err) == utils.ErrNotFound:
			// nothing to do
		default:
			return err
		}
	}

	err = delete("access keys", DefaultUserName, func(name Name) error {
		return asAdmin.DeleteAccessKeys(name, accessKeyID)
	})
	if err != nil {
		return err
	}

	err = delete("group", DefaultGroupName, asAdmin.DeleteGroup)
	if err != nil {
		return err
	}
	err = delete("user", DefaultUserName, asAdmin.DeleteUser)
	if err != nil {
		return err
	}
	if policy != nil {
		err := asAdmin.DeletePolicy(ARN(*policy.Arn))
		if err != nil {
			if errors.Cause(err) != utils.ErrNotFound {
				return err
			}
			log.Info("not found policy", "ARN", *policy.Arn)
		} else {
			log.Info("deleted policy", "ARN", *policy.Arn)
		}
	}

	// TODO clean up the Lambda functions and S3 objects

	return nil
}
