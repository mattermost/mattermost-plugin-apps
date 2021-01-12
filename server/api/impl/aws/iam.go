// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"
)

type policyDocument struct {
	Version   string
	Statement []policyStatementEntry
}

type policyStatementEntry struct {
	Effect   string
	Action   []string
	Resource string
}

type rolePolicyDocument struct {
	Version   string
	Statement []rolePolicyStatementEntry
}

type rolePolicyStatementEntry struct {
	Effect    string
	Action    string
	Principal rolePrincipal
}

type rolePrincipal struct {
	Service string
}

func (c *Client) makeLambdaFunctionDefaultPolicy() (string, error) {
	// Builds our policy document for IAM.
	policy := policyDocument{
		Version: "2012-10-17",
		Statement: []policyStatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents",
				},
				Resource: "*",
			},
		},
	}

	b, err := json.Marshal(&policy)
	if err != nil {
		return "", errors.Wrap(err, "can't marshal policy")
	}
	policyName := "my_cool_policy_name"

	arn := ""
	iamService := c.Service().iam
	out, err := iamService.CreatePolicy(&iam.CreatePolicyInput{
		PolicyDocument: aws.String(string(b)),
		PolicyName:     &policyName,
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if !ok || awsErr.Code() != iam.ErrCodeEntityAlreadyExistsException {
			return "", errors.Wrap(err, "can't create default lambda function policy")
		}

		if err = iamService.ListPoliciesPages(&iam.ListPoliciesInput{},
			func(page *iam.ListPoliciesOutput, lastPage bool) bool {
				for _, pol := range page.Policies {
					if *pol.PolicyName == policyName {
						arn = *pol.Arn
						return false
					}
				}
				return true
			},
		); err != nil {
			return "", errors.Wrap(err, "can't get policy arn")
		}
	} else {
		arn = *out.Policy.Arn
	}

	role, err := c.createRole(arn)
	if err != nil {
		return "", errors.Wrap(err, "can't create role")
	}
	return role, nil
}

func (c *Client) createRole(policyARN string) (string, error) {
	rolePolicy := rolePolicyDocument{
		Version: "2012-10-17",
		Statement: []rolePolicyStatementEntry{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: rolePrincipal{
					Service: "lambda.amazonaws.com",
				},
			},
		},
	}
	b, err := json.Marshal(&rolePolicy)
	if err != nil {
		return "", errors.Wrap(err, "can't marshal role policy")
	}
	roleName := "my_cool_role_name1"
	roleARN := ""
	iamService := c.Service().iam
	out, err := iamService.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(b)),
		RoleName:                 &roleName,
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if !ok || awsErr.Code() != iam.ErrCodeEntityAlreadyExistsException {
			return "", errors.Wrap(err, "can't create default lambda function role")
		}

		if err = iamService.ListRolesPages(&iam.ListRolesInput{},
			func(page *iam.ListRolesOutput, lastPage bool) bool {
				for _, r := range page.Roles {
					if *r.RoleName == roleName {
						roleARN = *r.Arn
						return false
					}
				}
				return true
			},
		); err != nil {
			return "", errors.Wrap(err, "can't get role arn")
		}
	} else {
		roleARN = *out.Role.Arn
		if _, err := iamService.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: &policyARN,
			RoleName:  &roleName,
		}); err != nil {
			return "", errors.Wrap(err, "can't attach role policy")
		}
	}
	return roleARN, nil
}
