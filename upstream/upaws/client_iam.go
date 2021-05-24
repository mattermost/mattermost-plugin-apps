// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (c *client) CreateUser(name Name) (ARN, error) {
	out, err := c.iam.CreateUser(&iam.CreateUserInput{
		UserName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrapf(err, "failed to create user %s", name)
	}
	return ARN(*out.User.Arn), nil
}

func (c *client) DeleteUser(name Name) error {
	_, err := c.iam.DeleteUser(&iam.DeleteUserInput{
		UserName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(name)
		}
		return errors.Wrapf(err, "failed to delete user %s", name)
	}
	return nil
}

func (c *client) FindUser(name Name) (ARN, error) {
	out, err := c.iam.GetUser(&iam.GetUserInput{
		UserName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return "", utils.NewNotFoundError(name)
		}
		return "", errors.Wrapf(err, "failed to find user %s", name)
	}
	return ARN(*out.User.Arn), nil
}

func (c *client) CreateGroup(name Name) (ARN, error) {
	out, err := c.iam.CreateGroup(&iam.CreateGroupInput{
		GroupName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrapf(err, "failed to create group %s", name)
	}
	return ARN(*out.Group.Arn), nil
}

func (c *client) DeleteGroup(name Name) error {
	_, err := c.iam.DeleteGroup(&iam.DeleteGroupInput{
		GroupName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(name)
		}
		return errors.Wrapf(err, "failed to delete group %s", name)
	}
	return nil
}

func (c *client) FindGroup(name Name) (ARN, error) {
	out, err := c.iam.GetGroup(&iam.GetGroupInput{
		GroupName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return "", utils.NewNotFoundError(name)
		}
		return "", errors.Wrapf(err, "failed to find group %s", name)
	}
	return ARN(*out.Group.Arn), nil
}

func (c *client) AddUserToGroup(u, g Name) error {
	_, err := c.iam.AddUserToGroup(&iam.AddUserToGroupInput{
		UserName:  u.AWSString(),
		GroupName: g.AWSString(),
	})
	return err
}

func (c *client) RemoveUserFromGroup(u, g Name) error {
	_, err := c.iam.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{
		UserName:  u.AWSString(),
		GroupName: g.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.ErrNotFound
		}
		return errors.Wrap(err, "failed to remove user from group")
	}
	return nil
}

func (c *client) AttachGroupPolicy(g Name, p ARN) error {
	_, err := c.iam.AttachGroupPolicy(&iam.AttachGroupPolicyInput{
		GroupName: g.AWSString(),
		PolicyArn: p.AWSString(),
	})
	return err
}

func (c *client) DetachGroupPolicy(g Name, p ARN) error {
	_, err := c.iam.DetachGroupPolicy(&iam.DetachGroupPolicyInput{
		GroupName: g.AWSString(),
		PolicyArn: p.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.ErrNotFound
		}
		return errors.Wrap(err, "failed to detach policy from group")
	}
	return nil
}

func (c *client) CreateAccessKey(u Name) (string, string, error) {
	out, err := c.iam.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: u.AWSString(),
	})
	if err != nil {
		return "", "", err
	}
	return *out.AccessKey.AccessKeyId, *out.AccessKey.SecretAccessKey, nil
}

func (c *client) DeleteAccessKeys(u Name, accessKeyID string) error {
	_, err := c.iam.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName:    u.AWSString(),
		AccessKeyId: aws.String(accessKeyID),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(u)
		}
		return errors.Wrapf(err, "failed to delete user access keys: %s", u)
	}
	return nil
}

func (c *client) CreatePolicy(name Name, data string) (ARN, error) {
	out, err := c.iam.CreatePolicy(&iam.CreatePolicyInput{
		PolicyName:     name.AWSString(),
		PolicyDocument: aws.String(data),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrapf(err, "failed to create policy %s", name)
	}

	return ARN(*out.Policy.Arn), nil
}

func (c *client) DeletePolicy(arn ARN) error {
	_, err := c.iam.DeletePolicy(&iam.DeletePolicyInput{
		PolicyArn: arn.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(arn)
		}
		return errors.Wrapf(err, "failed to delete policy %s", arn)
	}
	return nil
}

func (c *client) FindPolicy(policyName Name) (*iam.Policy, error) {
	var p *iam.Policy
	err := c.iam.ListPoliciesPages(&iam.ListPoliciesInput{}, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		for _, pol := range page.Policies {
			if *pol.PolicyName == string(policyName) {
				p = pol
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, errors.Wrapf(err, "can't find policy %s", policyName)
	}
	if p == nil {
		return nil, utils.NewNotFoundError(policyName)
	}
	return p, nil
}

func (c *client) getPolicyVersionDocument(p *iam.Policy) (string, *PolicyDocument, error) {
	out, err := c.iam.GetPolicyVersion(&iam.GetPolicyVersionInput{
		PolicyArn: p.Arn,
		VersionId: p.DefaultVersionId,
	})
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to find policy "+*p.Arn)
	}
	if out.PolicyVersion == nil {
		return "", nil, utils.NewNotFoundError(*p.Arn)
	}

	doc, err := url.QueryUnescape(*out.PolicyVersion.Document)
	if err != nil {
		return "", nil, errors.Wrap(err, "can't decode policy document"+*p.Arn)
	}

	resp := PolicyDocument{}
	err = json.Unmarshal([]byte(doc), &resp)
	if err != nil {
		return "", nil, errors.Wrap(err, "can't decode policy document"+*p.Arn)
	}
	return doc, &resp, nil
}

func (c *client) AddResourcesToPolicyDocument(p *iam.Policy, toAdd []ARN) (string, error) {
	outList, err := c.iam.ListPolicyVersions(&iam.ListPolicyVersionsInput{
		PolicyArn: p.Arn,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to list policy versions for %s", *p.Arn)
	}

	// Delete all versions except the defaultone, otherwise may hit the limit of
	// 5 versions.
	for _, v := range outList.Versions {
		if !*v.IsDefaultVersion {
			_, err = c.iam.DeletePolicyVersion(&iam.DeletePolicyVersionInput{
				PolicyArn: p.Arn,
				VersionId: v.VersionId,
			})
			if err != nil {
				return "", errors.Wrapf(err, "failed to delete an old policy version for %s", *p.Arn)
			}
		}
	}

	orig, doc, err := c.getPolicyVersionDocument(p)
	if err != nil {
		return "", err
	}

	var statement PolicyStatement
	found := -1
	for i, s := range doc.Statement {
		if s.Sid == "AllowLambda" {
			statement = s
			found = i
			break
		}
	}
	statement = DefaultAllowLambdaStatement(statement)

	changed := false
NEXT_ADD:
	for _, a := range toAdd {
		for _, r := range statement.Resource {
			if r == string(a) {
				continue NEXT_ADD
			}
		}

		statement.Resource = append(statement.Resource, string(a))
		changed = true
	}
	if !changed {
		// Nothing to do, already there.
		return orig, nil
	}

	if found < 0 {
		doc.Statement = append(doc.Statement, statement)
	} else {
		doc.Statement[found] = statement
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return "", errors.Wrapf(err, "failed to encode new policy version for %s", *p.Arn)
	}
	newDoc := string(data)

	_, err = c.iam.CreatePolicyVersion(&iam.CreatePolicyVersionInput{
		PolicyArn:      p.Arn,
		PolicyDocument: aws.String(newDoc),
		SetAsDefault:   aws.Bool(true),
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to create new policy version for %s", *p.Arn)
	}

	return newDoc, nil
}

func (c *client) FindRole(name Name) (ARN, error) {
	out, err := c.iam.GetRole(&iam.GetRoleInput{
		RoleName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return "", utils.NewNotFoundError(name)
		}
		return "", errors.Wrapf(err, "failed to find role %s", name)
	}
	return ARN(*out.Role.Arn), nil
}

func (c *client) CreateRole(name Name) (ARN, error) {
	out, err := c.iam.CreateRole(&iam.CreateRoleInput{
		RoleName:                 name.AWSString(),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			return "", utils.NewAlreadyExistsError(name)
		}
		return "", errors.Wrapf(err, "failed to create role %s", name)
	}
	return ARN(*out.Role.Arn), nil
}

func (c *client) DeleteRole(name Name) error {
	_, err := c.iam.DeleteRole(&iam.DeleteRoleInput{
		RoleName: name.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError(name)
		}
		return errors.Wrapf(err, "failed to delete role %s", name)
	}
	return nil
}

func (c *client) AttachRolePolicy(roleName Name, policyARN ARN) error {
	_, err := c.iam.AttachRolePolicy(&iam.AttachRolePolicyInput{
		RoleName:  roleName.AWSString(),
		PolicyArn: policyARN.AWSString(),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return utils.NewNotFoundError("role %s policy %s", roleName, policyARN)
		}
		return errors.Wrapf(err, "failed to attach role %s to policy %s", roleName, policyARN)
	}
	return nil
}
