package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
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
	Short: "A tool to manage Mattermost Apps.",
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

func makeTestUpstream() (*upaws.Upstream, error) {
	region := os.Getenv(upaws.RegionEnvVar)
	if region == "" {
		return nil, errors.Errorf("no AWS region was provided. Please set %s", upaws.RegionEnvVar)
	}
	accessKey := os.Getenv(upaws.AccessEnvVar)
	if accessKey == "" {
		return nil, errors.Errorf("no AWS access key was provided. Please set %s", upaws.AccessEnvVar)
	}
	secretKey := os.Getenv(upaws.SecretEnvVar)
	if secretKey == "" {
		return nil, errors.Errorf("no AWS secret key was provided. Please set %s", upaws.SecretEnvVar)
	}

	log.Debug("Using AWS credentials", "AccessKeyID", utils.LastN(accessKey, 7), "AccessKeySecretID", utils.LastN(secretKey, 4))
	return upaws.MakeUpstream(accessKey, secretKey, region, upaws.S3BucketName(), &log)
}

func makeProvisionClient() (upaws.Client, error) {
	region := os.Getenv(upaws.RegionEnvVar)
	if region == "" {
		return nil, errors.Errorf("no AWS region was provided. Please set %s", upaws.RegionEnvVar)
	}
	accessKey := os.Getenv(upaws.ProvisionAccessEnvVar)
	if accessKey == "" {
		return nil, errors.Errorf("no AWS access key was provided. Please set %s", upaws.ProvisionAccessEnvVar)
	}
	secretKey := os.Getenv(upaws.ProvisionSecretEnvVar)
	if secretKey == "" {
		return nil, errors.Errorf("no AWS secret key was provided. Please set %s", upaws.ProvisionSecretEnvVar)
	}

	log.Debug("Using AWS credentials", "AccessKeyID", utils.LastN(accessKey, 7), "AccessKeySecretID", utils.LastN(secretKey, 4))
	return upaws.MakeClient(accessKey, secretKey, region, &log, "Provisioner (appsctl)")
}
