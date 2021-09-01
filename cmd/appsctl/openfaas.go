package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upopenfaas"
)

func init() {
	rootCmd.AddCommand(openfaasCmd)

	// provision
	openfaasCmd.AddCommand(openfaasProvisionCmd)
	openfaasProvisionCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with caution in production.")

	// test
	// openfaasCmd.AddCommand(openfaasTestCmd)
}

var openfaasCmd = &cobra.Command{
	Use:   "openfaas",
	Short: "Provision Mattermost Apps to OpenFaaS or faasd",
}

var openfaasProvisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision a Mattermost app to OpenFaaS or faasd",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]
		gateway := os.Getenv(upopenfaas.EnvGatewayURL)

		m, err := upopenfaas.ProvisionApp(bundlePath, log, shouldUpdate, gateway)
		if err != nil {
			return err
		}
		if m.OpenFAAS == nil || len(m.OpenFAAS.Functions) == 0 {
			return errors.New("no functions to provision, check manifest.json")
		}

		rootURL, err := upopenfaas.RootURL(*m, gateway, "/")
		if err != nil {
			return err
		}

		fmt.Printf("\nProvisioned '%s' to OpenFaaS.\n", m.DisplayName)
		fmt.Printf("You can install it now in Mattermost using:\n")
		fmt.Printf("  /apps install url %s/%s\n\n", rootURL, "manifest.json")
		return nil
	},
}

// func helloKubeless() apps.App {
// 	return apps.App{
// 		DeployType: apps.DeployKubeless,
// 		Manifest: apps.Manifest{
// 			AppID:   "hello-openfaas",
// 			Version: "demo",
// 			Kubeless: &apps.Kubeless{
// 				Functions: []apps.KubelessFunction{
// 					{
// 						Path: "/",
// 						Runtime:  "nodejs14", // see /examples/js/hello-world
// 						Handler:  "app.handler",
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// var openfaasTestCmd = &cobra.Command{
// 	Use:   "test",
// 	Short: "provisions and tests 'hello-lambda'",
// 	Long: `Test commands us the 'hello-lambda' example app for testing, see
// https://github.com/mattermost/mattermost-plugin-apps/tree/master/examples/go/hello-lambda/README.md

// The App needs to be built with 'make dist' in its own directory, then use
// `,
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		upTest, err := upopenfaas.MakeUpstream()
// 		if err != nil {
// 			return err
// 		}

// 		app := helloKubeless()
// 		creq := apps.CallRequest{
// 			Call: apps.Call{
// 				Path: "/ping",
// 			},
// 		}
// 		log.Debugw("Invoking test function",
// 			"app_id", app.AppID,
// 			"version", app.Version,
// 			"path", creq.Call.Path,
// 			"handler", app.Manifest.Kubeless.Functions[0].Handler)
// 		resp, err := upTest.Roundtrip(app, creq, false)
// 		if err != nil {
// 			return err
// 		}
// 		defer resp.Close()

// 		data, err := io.ReadAll(resp)
// 		if err != nil {
// 			return err
// 		}
// 		log.Debugf("Received: %s", string(data))

// 		cresp := apps.CallResponse{}
// 		_ = json.Unmarshal(data, &cresp)
// 		expected := apps.CallResponse{Markdown: "PONG", Type: apps.CallResponseTypeOK}
// 		if cresp != expected {
// 			return errors.Errorf("invalid value received: %s", string(data))
// 		}

// 		fmt.Println("OK")
// 		return nil
// 	},
// }
