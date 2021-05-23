package main

import (
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
)

var (
	shouldUpdate      bool
	invokePolicyName  string
	executePolicyName string
)

func init() {
	rootCmd.AddCommand(
		provisionCmd,
	)

	provisionCmd.AddCommand(
		provisionAppCmd,
	)

	provisionAppCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with causion in production.")
	provisionAppCmd.Flags().StringVar(&invokePolicyName, "invoke-policy", upaws.DefaultPolicyName, "name of the policy used to invoke Apps on AWS.")
	provisionAppCmd.Flags().StringVar(&executePolicyName, "execute-policy", "TODO", "TODO.")
}

var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision aws resources",
}

var provisionAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Provision a Mattermost app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		awsClient, err := createAWSClient(false)
		if err != nil {
			return err
		}

		bucket := upaws.S3BucketName()
		err = upaws.ProvisionAppFromFile(awsClient, bucket, executePolicyName, invokePolicyName, args[0], shouldUpdate, &log)
		if err != nil {
			return err
		}

		return nil
	},
}
