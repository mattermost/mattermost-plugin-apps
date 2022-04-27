// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func CleanAWS(asAdmin Client, accessKeyID string, log utils.Logger) error {
	delete := func(typ string, name Name, del func(Name) error) error {
		err := del(name)
		if err != nil {
			if !errors.Is(err, utils.ErrNotFound) {
				return err
			}
			log.Infow("not found "+typ, "key", name)
		} else {
			log.Infow("deleted "+typ, "key", name)
		}
		return nil
	}

	err := asAdmin.RemoveUserFromGroup(DefaultUserName, DefaultGroupName)
	switch {
	case err == nil:
		log.Infow("removed user from group", "user", DefaultUserName, "group", DefaultGroupName)
	case errors.Is(err, utils.ErrNotFound):
		// nothing to do
	default:
		return err
	}

	policy, err := asAdmin.FindPolicy(DefaultPolicyName)
	if err == nil {
		err = asAdmin.DetachGroupPolicy(DefaultGroupName, ARN(*policy.Arn))
		switch {
		case err == nil:
			log.Infow("detached policy from group", "policy", DefaultPolicyName, "group", DefaultGroupName)
		case errors.Is(err, utils.ErrNotFound):
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
			if !errors.Is(err, utils.ErrNotFound) {
				return err
			}
			log.Infow("not found policy", "ARN", *policy.Arn)
		} else {
			log.Infow("deleted policy", "ARN", *policy.Arn)
		}
	}

	// TODO clean up the Lambda functions and S3 objects

	return nil
}
