// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	provisionCmd.AddCommand(
		generateTerraformCmd,
	)
}

var generateTerraformCmd = &cobra.Command{
	Use:   "generate-terraform-data",
	Short: "Generate data for terraform to provision aws apps",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := aws.GetProvisionDataFromFile(args[0], &log)
		if err != nil {
			return errors.Wrap(err, "can't get provision data")
		}

		bytes, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return errors.Wrap(err, "can't marshal data")
		}
		cmd.Println(string(bytes))
		return nil
	},
}
