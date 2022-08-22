package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var (
	verbose bool
	quiet   bool
)

var log = utils.MustMakeCommandLogger(zapcore.InfoLevel)

var (
	dockerRegistry        string
	executeRoleName       string
	groupName             string
	install               bool
	invokePolicyName      string
	policyName            string
	shouldCreate          bool
	shouldCreateAccessKey bool
	shouldUpdate          bool
	userName              string
	environment           map[string]string
)

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
