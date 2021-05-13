package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps/awsapps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

var (
	verbose bool
	quiet   bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose (debug) output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet (errors only) output")
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
		log.SetLevel(logrus.InfoLevel)
		if verbose {
			log.SetLevel(logrus.DebugLevel)
		}
		if quiet {
			log.SetLevel(logrus.ErrorLevel)
		}
	},
}

func createAWSClient() (awsclient.Client, error) {
	region := os.Getenv(awsapps.RegionEnvVar)
	if region == "" {
		return nil, errors.Errorf("no AWS region was provided. Please set %s", awsapps.RegionEnvVar)
	}
	accessKey := os.Getenv(awsapps.ProvisionAccessEnvVar)
	if accessKey == "" {
		return nil, errors.Errorf("no AWS access key was provided. Please set %s", awsapps.ProvisionAccessEnvVar)
	}
	secretKey := os.Getenv(awsapps.ProvisionSecretEnvVar)
	if secretKey == "" {
		return nil, errors.Errorf("no AWS secret key was provided. Please set %s", awsapps.ProvisionSecretEnvVar)
	}

	return awsclient.MakeClient(accessKey, secretKey, region, &log)
}
