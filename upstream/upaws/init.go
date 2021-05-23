// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/awsclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const LambdaExecutionPolicy = `arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole`

const (
	DefaultPolicyName = "mattermost-apps-invoke-policy"
	DefaultUserName   = "mattermost-apps-invoke"
	DefaultGroupName  = "mattermost-apps-invoke-group"
)

type InitParams struct {
	Bucket                string
	Policy                string
	User                  string
	Group                 string
	ShouldCreate          bool
	ShouldCreateAccessKey bool
}

type InitResult struct {
	Bucket          string
	PolicyARN       string
	UserARN         string
	GroupARN        string
	AccessKeyID     string
	AccessKeySecret string
}

func InitApps(asAdmin awsclient.Client, params InitParams, log Logger) (r *InitResult, err error) {
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

	ensure := func(typ, name string, find, create func(string) (string, error)) (string, error) {
		var arn string
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

	r.PolicyARN, err = ensure("policy", params.Policy,
		func(name string) (string, error) {
			var p *iam.Policy
			p, err = asAdmin.FindPolicy(name)
			if err != nil {
				return "", err
			}
			return *p.Arn, nil
		},
		func(name string) (string, error) {
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

	err = asAdmin.AttachGroupPolicy(params.Group, r.PolicyARN)
	if err != nil {
		return nil, err
	}
	err = asAdmin.AddUserToGroup(params.User, params.Group)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func CleanApps(asAdmin awsclient.Client, log Logger) error {
	delete := func(typ, name string, del func(string) error) error {
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
		err = asAdmin.DetachGroupPolicy(DefaultGroupName, *policy.Arn)
		switch {
		case err == nil:
			log.Info("detached policy from group", "policy", DefaultPolicyName, "group", DefaultGroupName)
		case errors.Cause(err) == utils.ErrNotFound:
			// nothing to do
		default:
			return err
		}
	}

	err = delete("access keys", DefaultUserName, asAdmin.DeleteAccessKeys)
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
		err = delete("policy", *policy.Arn, asAdmin.DeletePolicy)
		if err != nil {
			return err
		}
	}

	// TODO clean up the Lambda functions and S3 objects

	return nil
}
