package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
