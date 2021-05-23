// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsclient

import (
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (c *client) CreateUser(name string) (string, error) {
	out, err := c.iam.CreateUser(&iam.CreateUserInput{
		UserName: aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrap(err, "failed to create user "+name)
	}
	return *out.User.Arn, nil
}

func (c *client) DeleteUser(name string) error {
	_, err := c.iam.DeleteUser(&iam.DeleteUserInput{
		UserName: aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(name)
		}
		return errors.Wrap(err, "failed to delete user "+name)
	}
	return nil
}

func (c *client) FindUser(name string) (string, error) {
	out, err := c.iam.GetUser(&iam.GetUserInput{
		UserName: aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return "", utils.NewNotFoundError(name)
		}
		return "", errors.Wrap(err, "failed to find user "+name)
	}
	return *out.User.Arn, nil
}

func (c *client) CreateGroup(name string) (string, error) {
	out, err := c.iam.CreateGroup(&iam.CreateGroupInput{
		GroupName: aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrap(err, "failed to create group "+name)
	}
	return *out.Group.Arn, nil
}

func (c *client) DeleteGroup(name string) error {
	_, err := c.iam.DeleteGroup(&iam.DeleteGroupInput{
		GroupName: aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(name)
		}
		return errors.Wrap(err, "failed to delete group "+name)
	}
	return nil
}

func (c *client) FindGroup(name string) (string, error) {
	out, err := c.iam.GetGroup(&iam.GetGroupInput{
		GroupName: aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return "", utils.NewNotFoundError(name)
		}
		return "", errors.Wrap(err, "failed to find group "+name)
	}
	return *out.Group.Arn, nil
}

func (c *client) AddUserToGroup(u, g string) error {
	_, err := c.iam.AddUserToGroup(&iam.AddUserToGroupInput{
		UserName:  aws.String(u),
		GroupName: aws.String(g),
	})
	return err
}

func (c *client) RemoveUserFromGroup(u, g string) error {
	_, err := c.iam.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{
		UserName:  aws.String(u),
		GroupName: aws.String(g),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.ErrNotFound
		}
		return errors.Wrap(err, "failed to remove user from group")
	}
	return nil
}

func (c *client) AttachGroupPolicy(g, p string) error {
	_, err := c.iam.AttachGroupPolicy(&iam.AttachGroupPolicyInput{
		GroupName: aws.String(g),
		PolicyArn: aws.String(p),
	})
	return err
}

func (c *client) DetachGroupPolicy(g, p string) error {
	_, err := c.iam.DetachGroupPolicy(&iam.DetachGroupPolicyInput{
		GroupName: aws.String(g),
		PolicyArn: aws.String(p),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.ErrNotFound
		}
		return errors.Wrap(err, "failed to detach policy from group")
	}
	return nil
}

func (c *client) CreateAccessKey(u string) (string, string, error) {
	out, err := c.iam.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(u),
	})
	if err != nil {
		return "", "", err
	}
	return *out.AccessKey.AccessKeyId, *out.AccessKey.SecretAccessKey, nil
}

func (c *client) DeleteAccessKeys(u string) error {
	_, err := c.iam.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName: aws.String(u),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(u)
		}
		return errors.Wrap(err, "failed to delete user access keys: "+u)
	}
	return nil
}

func (c *client) CreatePolicy(name, data string) (string, error) {
	out, err := c.iam.CreatePolicy(&iam.CreatePolicyInput{
		PolicyDocument: aws.String(data),
		PolicyName:     aws.String(name),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrap(err, "failed to create policy "+name)
	}

	return *out.Policy.Arn, nil
}

func (c *client) DeletePolicy(arn string) error {
	_, err := c.iam.DeletePolicy(&iam.DeletePolicyInput{
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok || awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(arn)
		}
		return errors.Wrap(err, "failed to delete policy "+arn)
	}
	return nil
}

func (c *client) FindPolicy(policyName string) (*iam.Policy, error) {
	var p *iam.Policy
	err := c.iam.ListPoliciesPages(&iam.ListPoliciesInput{}, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		for _, pol := range page.Policies {
			if *pol.PolicyName == policyName {
				p = pol
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, errors.Wrap(err, "can't find policy "+policyName)
	}
	if p == nil {
		return nil, utils.NewNotFoundError(policyName)
	}
	return p, nil
}

func (c *client) GetPolicyVersionDocument(p *iam.Policy) (map[string]interface{}, error) {
	out, err := c.iam.GetPolicyVersion(&iam.GetPolicyVersionInput{
		PolicyArn: p.Arn,
		VersionId: p.DefaultVersionId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to find policy "+*p.Arn)
	}
	if out.PolicyVersion == nil {
		return nil, utils.NewNotFoundError(*p.Arn)
	}

	doc, err := url.QueryUnescape(*out.PolicyVersion.Document)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode policy document"+*p.Arn)
	}

	resp := map[string]interface{}{}
	err = json.Unmarshal([]byte(doc), &resp)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode policy document"+*p.Arn)
	}
	return resp, nil
}

func (c *client) ListPolicyVersions(policyARN string) ([]*iam.PolicyVersion, error) {
	out, err := c.iam.ListPolicyVersions(&iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyARN),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list policy versions for "+policyARN)
	}
	if len(out.Versions) == 0 {
		return nil, utils.NewNotFoundError(policyARN)
	}
	return out.Versions, nil
}
