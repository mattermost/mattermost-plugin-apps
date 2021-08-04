package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var (
	verbose bool
	quiet   bool
)

var log = utils.MustMakeCommandLogger(zapcore.InfoLevel)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose (debug) output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet (errors only) output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatalf("command failed")
	}
}

var rootCmd = &cobra.Command{
	Use:   "appsctl",
	Short: "A tool to manage Mattermost Apps.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log = utils.MustMakeCommandLogger(zapcore.DebugLevel)
		}
		if quiet {
			log = utils.MustMakeCommandLogger(zapcore.ErrorLevel)
		}
	},
}

func AsProvisioner() (upaws.Client, error) {
	return createClient(false)
}

func AsTest() (upaws.Client, error) {
	return createClient(true)
}

func createClient(invoke bool) (upaws.Client, error) {
	accessVar, secretVar := upaws.ProvisionAccessEnvVar, upaws.ProvisionSecretEnvVar
	if invoke {
		accessVar, secretVar = upaws.AccessEnvVar, upaws.SecretEnvVar
	}

	region := os.Getenv(upaws.RegionEnvVar)
	if region == "" {
		return nil, errors.Errorf("no AWS region was provided. Please set %s", upaws.RegionEnvVar)
	}
	accessKey := os.Getenv(accessVar)
	if accessKey == "" {
		return nil, errors.Errorf("no AWS access key was provided. Please set %s", accessVar)
	}
	secretKey := os.Getenv(secretVar)
	if secretKey == "" {
		return nil, errors.Errorf("no AWS secret key was provided. Please set %s", secretVar)
	}

	log.Debugw("Using AWS credentials", "AccessKeyID", utils.LastN(accessKey, 7), "AccessKeySecretID", utils.LastN(secretKey, 4))
	return upaws.MakeClient(accessKey, secretKey, region, log)
}
