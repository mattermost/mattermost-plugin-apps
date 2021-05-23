package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
)

var initParams upaws.InitParams
var shouldClean bool

func init() {
	rootCmd.AddCommand(
		initCmd,
	)

	initCmd.AddCommand(initAWSCmd)
	initAWSCmd.Flags().BoolVar(&shouldClean, "clean", false, "Destroy all default AWS resources, including the S3 bucket.")
	initAWSCmd.Flags().BoolVar(&initParams.ShouldCreate, "create", false, "Create resources (user, group, policy, bucket) that don't already exist, using the default configuration.")
	initAWSCmd.Flags().BoolVar(&initParams.ShouldCreateAccessKey, "create-access-key", false, "Create new access key for the user (or you can safely do it in AWS Console).")
	initAWSCmd.Flags().StringVar(&initParams.User, "user", upaws.DefaultUserName, "Username to use for invoking the AWS App from Mattermost Server.")
	initAWSCmd.Flags().StringVar(&initParams.Policy, "policy", upaws.DefaultPolicyName, "Name of the policy to control access of AWS services directly by Mattermost Server (user).")
	initAWSCmd.Flags().StringVar(&initParams.Group, "group", upaws.DefaultGroupName, "Name of the user group connecting the invoking user to the invoke policy.")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize upstream hosting services (e.g. AWS) for deploying Mattermost Apps",
}

var initAWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "Initialize AWS to deploy Mattermost Apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		asAdmin, err := createAWSClient(false)
		if err != nil {
			return err
		}

		initParams.Bucket = upaws.S3BucketName()

		if shouldClean {
			return upaws.CleanApps(asAdmin, &log)
		}

		out, err := upaws.InitApps(asAdmin, initParams, &log)
		if err != nil {
			return err
		}

		fmt.Printf("Ready to deploy AWS Lambda Mattermost Apps!\n\n")

		fmt.Printf("User:\t%q\n", out.UserARN)
		fmt.Printf("Group:\t%q\n", out.GroupARN)
		fmt.Printf("Policy:\t%q\n", out.PolicyARN)
		fmt.Printf("Bucket:\t%q\n", out.Bucket)

		if initParams.ShouldCreateAccessKey {
			fmt.Printf("\nPlease store the Access Key securely, it will not be viewable again.\n\n")
			fmt.Printf("Access Key ID:\t%s\n", out.AccessKeyID)
			fmt.Printf("Access Key Secret:\t%s\n", out.AccessKeySecret)
		}

		return nil
	},
}
