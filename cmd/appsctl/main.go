package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
)

var (
	verbose bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("command failed")
	}
}

var rootCmd = &cobra.Command{
	Use:   "appsctl",
	Short: "A tool to manage Mattermost apps.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(logrus.DebugLevel)
		}
	},
}

func createAWSClient() (*aws.Client, error) {
	accessKey := os.Getenv("APPS_PROVISION_AWS_ACCESS_KEY")
	secretKey := os.Getenv("APPS_PROVISION_AWS_SECRET_KEY")

	if accessKey == "" {
		return nil, errors.New("no AWS access key was provided. Please set APPS_PROVISION_AWS_ACCESS_KEY")
	}

	if secretKey == "" {
		return nil, errors.New("no AWS secret key was provided. Please set APPS_PROVISION_AWS_SECRET_KEY")
	}

	return aws.NewAWSClient(accessKey, secretKey, &log), nil
}
