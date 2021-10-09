package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upopenfaas"
)

func init() {
	rootCmd.AddCommand(openfaasCmd)

	// deploy
	openfaasCmd.AddCommand(openfaasDeployCmd)
	openfaasDeployCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with caution in production.")
	openfaasDeployCmd.Flags().BoolVar(&install, "install", false, "Install the deployed App to Mattermost")
	openfaasDeployCmd.Flags().StringVar(&dockerRegistry, "docker-registry", "", "Docker image prefix, usually the docker registry to use for deploying functions.")

	// test
	// openfaasCmd.AddCommand(openfaasTestCmd)
}

var openfaasCmd = &cobra.Command{
	Use:   "openfaas",
	Short: "Deploy Mattermost Apps to OpenFaaS or faasd",
}

var openfaasDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Mattermost app to OpenFaaS or faasd",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]
		gateway := os.Getenv(upopenfaas.EnvGatewayURL)

		m, err := upopenfaas.DeployApp(bundlePath, log, shouldUpdate, gateway, dockerRegistry)
		if err != nil {
			return err
		}
		if m.OpenFAAS == nil || len(m.OpenFAAS.Functions) == 0 {
			return errors.New("no functions to deploy, check manifest.json")
		}

		if err = updateMattermost(*m, apps.DeployOpenFAAS, install); err != nil {
			return err
		}

		fmt.Printf("\nDeployed '%s' to OpenFaaS.\n", m.DisplayName)

		if !install {
			fmt.Printf("You can now install it in Mattermost using:\n")
			fmt.Printf("  /apps install listed %s\n\n", m.AppID)
		}
		return nil
	},
}
