// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
)

func init() {
	rootCmd.AddCommand(
		terraformCmd,
	)

	terraformCmd.AddCommand(terraformGenerateCmd)
}

var terraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Generate Terraform data for Mattermost Cloud",
}

var terraformGenerateCmd = &cobra.Command{
	Use:   "generate-data",
	Short: "Generate Terraform data for Mattermost Cloud",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := upaws.GetProvisionDataFromFile(args[0], log)
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
