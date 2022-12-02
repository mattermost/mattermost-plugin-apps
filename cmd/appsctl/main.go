package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	root "github.com/mattermost/mattermost-plugin-apps"
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
	_ = rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:     "appsctl",
	Short:   "A tool to manage Mattermost Apps.",
	Version: root.Manifest.Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log = utils.MustMakeCommandLogger(zapcore.DebugLevel)
		}
		if quiet {
			log = utils.MustMakeCommandLogger(zapcore.ErrorLevel)
		}
	},
}
