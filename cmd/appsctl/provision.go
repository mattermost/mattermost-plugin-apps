package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
)

var (
	shouldUpdate     bool
	invokePolicyName string
	executeRoleName  string
)

func init() {
	rootCmd.AddCommand(
		provisionCmd,
	)

	provisionCmd.AddCommand(
		provisionAppCmd,
	)

	provisionAppCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with causion in production.")
	provisionAppCmd.Flags().StringVar(&invokePolicyName, "policy", upaws.DefaultPolicyName, "name of the policy used to invoke Apps on AWS.")
	provisionAppCmd.Flags().StringVar(&executeRoleName, "execute-role", upaws.DefaultExecuteRoleName, "name of the role to be assumed by running Lambdas.")
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
		out, err := upaws.ProvisionAppFromFile(awsClient, args[0], &log, upaws.ProvisionAppParams{
			Bucket:           bucket,
			InvokePolicyName: upaws.Name(invokePolicyName),
			ExecuteRoleName:  upaws.Name(executeRoleName),
			ShouldUpdate:     shouldUpdate,
		})
		if err != nil {
			return err
		}

		fmt.Printf("<><> %+v\n", out)

		return nil
	},
}
