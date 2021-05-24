package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
)

var shouldClean bool
var shouldCreate bool
var shouldCreateAccessKey bool
var userName string
var policyName string
var groupName string

func init() {
	rootCmd.AddCommand(
		initCmd,
	)

	initCmd.AddCommand(initAWSCmd)
	initAWSCmd.Flags().BoolVar(&shouldClean, "clean", false, "Destroy all default AWS resources, including the S3 bucket.")
	initAWSCmd.Flags().BoolVar(&shouldCreate, "create", false, "Create resources (user, group, policy, bucket) that don't already exist, using the default configuration.")
	initAWSCmd.Flags().BoolVar(&shouldCreateAccessKey, "create-access-key", false, "Create new access key for the user (or you can safely do it in AWS Console).")
	initAWSCmd.Flags().StringVar(&userName, "user", upaws.DefaultUserName, "Username to use for invoking the AWS App from Mattermost Server.")
	initAWSCmd.Flags().StringVar(&policyName, "policy", upaws.DefaultPolicyName, "Name of the policy to control access of AWS services directly by Mattermost Server (user).")
	initAWSCmd.Flags().StringVar(&groupName, "group", upaws.DefaultGroupName, "Name of the user group connecting the invoking user to the invoke policy.")
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

		if shouldClean {
			accessKeyID := os.Getenv(upaws.AccessEnvVar)
			if accessKeyID == "" {
				return errors.Errorf("no AWS access key was provided. Please set %s", upaws.AccessEnvVar)
			}

			return upaws.CleanApps(asAdmin, accessKeyID, &log)
		}

		out, err := upaws.InitApps(asAdmin, &log, upaws.InitParams{
			Bucket:                upaws.S3BucketName(),
			User:                  upaws.Name(userName),
			Group:                 upaws.Name(groupName),
			Policy:                upaws.Name(policyName),
			ExecuteRole:           upaws.Name(executeRoleName),
			ShouldCreate:          shouldCreate,
			ShouldCreateAccessKey: shouldCreateAccessKey,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Ready to deploy AWS Lambda Mattermost Apps!\n\n")

		fmt.Printf("User:\t%q\n", out.UserARN)
		fmt.Printf("Group:\t%q\n", out.GroupARN)
		fmt.Printf("Policy:\t%q\n", out.PolicyARN)
		fmt.Printf("Bucket:\t%q\n", out.Bucket)

		if shouldCreateAccessKey {
			fmt.Printf("\nPlease store the Access Key securely, it will not be viewable again.\n\n")
			fmt.Printf("export %s='%s'\n", upaws.AccessEnvVar, out.AccessKeyID)
			fmt.Printf("export %s='%s'\n", upaws.SecretEnvVar, out.AccessKeySecret)
		}

		return nil
	},
}
