//go:build ignore

// Kubeless is not longer supported: https://mattermost.atlassian.net/browse/MM-40011

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
)

func init() {
	rootCmd.AddCommand(kubelessCmd)

	// deploy
	kubelessCmd.AddCommand(kubelessDeployCmd)
	kubelessDeployCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with caution in production.")
	kubelessDeployCmd.Flags().BoolVar(&install, "install", false, "Install the deployed App to Mattermost")

	// test
	kubelessCmd.AddCommand(kubelessTestCmd)
}

var kubelessCmd = &cobra.Command{
	Use:   "kubeless",
	Short: "Deploy Mattermost Apps to Kubeless",
}

var kubelessDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Mattermost app to Kubeless",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]

		m, err := upkubeless.DeployApp(bundlePath, log, shouldUpdate)
		if err != nil {
			return err
		}
		if !m.Contains(apps.DeployKubeless) {
			return errors.New("manifest.json unexpectedly contains no Kubeless data")
		}
		if len(m.Kubeless.Functions) == 0 {
			return errors.New("no Kubeless functions to deploy in manifest.json")
		}

		if err = updateMattermost(*m, apps.DeployKubeless, install); err != nil {
			return err
		}

		fmt.Printf("\nDeployed '%s' to Kubeless, %v functions deployed.\n", m.DisplayName, len(m.Kubeless.Functions))

		if !install {
			fmt.Printf("You can now install it in Mattermost using:\n")
			fmt.Printf("  /apps install listed %s\n\n", m.AppID)
		}
		return nil
	},
}

func helloKubeless() apps.App {
	return apps.App{
		DeployType: apps.DeployKubeless,
		Manifest: apps.Manifest{
			AppID:   "hello-kubeless",
			Version: "0.8.0",
			Deploy: apps.Deploy{
				Kubeless: &apps.Kubeless{
					Functions: []apps.KubelessFunction{
						{
							Path:    "/",
							Runtime: "nodejs14", // see /examples/js/hello-world
							Handler: "app.handler",
						},
					},
				},
			},
		},
	}
}

var kubelessTestCmd = &cobra.Command{
	Use:   "test",
	Short: "deploys and tests 'hello-lambda'",
	Long: `Test commands us the 'hello-lambda' example app for testing, see
https://github.com/mattermost/mattermost-plugin-apps/tree/master/examples/go/hello-lambda/README.md

The App needs to be built with 'make dist' in its own directory, then use
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		upTest, err := upkubeless.MakeUpstream()
		if err != nil {
			return err
		}

		app := helloKubeless()
		creq := apps.CallRequest{
			Call: apps.Call{
				Path: "/ping",
			},
		}
		log.Debugw("Invoking test function",
			"app_id", app.AppID,
			"version", app.Version,
			"path", creq.Call.Path,
			"handler", app.Manifest.Kubeless.Functions[0].Handler)

		ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
		defer cancel()

		resp, err := upTest.Roundtrip(ctx, app, creq, false)
		if err != nil {
			return err
		}
		defer resp.Close()

		data, err := io.ReadAll(resp)
		if err != nil {
			return err
		}
		log.Debugf("Received: %s", string(data))

		cresp := apps.CallResponse{}
		_ = json.Unmarshal(data, &cresp)
		expected := apps.NewTextResponse("PONG")
		if cresp != expected {
			return errors.Errorf("invalid value received: %s", string(data))
		}

		fmt.Println("OK")
		return nil
	},
}
