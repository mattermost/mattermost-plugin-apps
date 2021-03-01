package main

import (
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

var (
	shouldUpdate bool
)

func init() {
	rootCmd.AddCommand(
		provisionCmd,
	)

	provisionCmd.AddCommand(
		provisionAppCmd,
		provisionBucketCmd,
	)

	provisionAppCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with causion in production.")
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
		awsClient, err := createAWSClient()
		if err != nil {
			return err
		}

		err = ProvisionAppFromFile(awsClient, args[0], shouldUpdate)
		if err != nil {
			return err
		}

		return nil
	},
}

var provisionBucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "Provision the central s3 bucket used to store app data",
	RunE: func(cmd *cobra.Command, args []string) error {
		awsClient, err := createAWSClient()
		if err != nil {
			return err
		}

		name := awsclient.BucketWithDefaults("")

		exists, err := awsClient.S3BucketExists(name)
		if err != nil {
			return err
		}

		if exists {
			log.Infof("Bucket %v already exists", name)
			return nil
		}

		err = awsClient.CreateS3Bucket(name)
		if err != nil {
			return err
		}

		log.Infof("Created bucket %s", name)

		return nil
	},
}
