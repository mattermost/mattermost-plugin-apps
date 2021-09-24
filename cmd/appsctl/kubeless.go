package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
)

func init() {
	rootCmd.AddCommand(kubelessCmd)

	// provision
	kubelessCmd.AddCommand(kubelessProvisionCmd)
	kubelessProvisionCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with caution in production.")
	kubelessProvisionCmd.Flags().BoolVar(&install, "install", false, "Install the deployed App to Mattermost")

	// test
	kubelessCmd.AddCommand(kubelessTestCmd)
}

var kubelessCmd = &cobra.Command{
	Use:   "kubeless",
	Short: "Provision Mattermost Apps to Kubeless",
}

var kubelessProvisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision a Mattermost app to Kubeless",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]

		m, err := upkubeless.ProvisionApp(bundlePath, log, shouldUpdate)
		if err != nil {
			return err
		}
		if m.Kubeless == nil || len(m.Kubeless.Functions) == 0 {
			return errors.New("no functions to provision, check manifest.json")
		}

		if err = updateMattermost(*m, apps.DeployKubeless, install); err != nil {
			return err
		}

		fmt.Printf("\nProvisioned '%s' to Kubeless, %v functions deployed.\n", m.DisplayName, len(m.Kubeless.Functions))

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
			Version: "demo",
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
	Short: "provisions and tests 'hello-lambda'",
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
		resp, err := upTest.Roundtrip(app, creq, false)
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
		expected := apps.NewOKResponse(nil, "PONG")
		if cresp != expected {
			return errors.Errorf("invalid value received: %s", string(data))
		}

		fmt.Println("OK")
		return nil
	},
}
