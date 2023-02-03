package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func init() {
	rootCmd.AddCommand(pluginCmd)

	// deploy
	pluginCmd.AddCommand(pluginDeployCmd)
	pluginDeployCmd.Flags().BoolVar(&install, "install", false, "Install the plugin as a Mattermost App")
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Deploy compatible Mattermost plugins as Apps",
}

var pluginDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a compatible Mattermost plugin as an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]

		appClient, err := getMattermostClient()
		if err != nil {
			return err
		}

		m, err := installPlugin(appClient, bundlePath)
		if err != nil {
			return err
		}

		if err = updateMattermost(appClient, *m, apps.DeployPlugin, install); err != nil {
			return err
		}

		fmt.Printf("\nDeployed '%s' as plugin.\n", m.DisplayName)

		if !install {
			fmt.Printf("You can now install it in Mattermost using:\n")
			fmt.Printf("  /apps install listed %s\n\n", m.AppID)
		}
		return nil
	},
}
