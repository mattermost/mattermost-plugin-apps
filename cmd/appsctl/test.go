package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
)

func init() {
	rootCmd.AddCommand(
		testCmd,
	)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test accessing a provisioned resource",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		awsClient, err := createAWSClient(true)
		if err != nil {
			return err
		}

		name := "test"
		if len(args) > 0 {
			name = args[0]
		}

		app := &apps.App{
			Manifest: apps.Manifest{
				AppID:   "test",
				Version: "vvvv",
				// AWSLambda: []apps.AWSLambda{
				// 	{
				// 		Path: "/",
				// 		Name: name,
				// 	},
				// },
			},
		}
		up := upaws.NewUpstream(app, awsClient, "")
		cr := &apps.CallRequest{
			Call: apps.Call{
				Path: "/test",
			},
		}
		resp, err := up.InvokeFunction(name, cr, false)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(resp)
		if err != nil {
			return err
		}
		log.Infof("received: %s", string(data))
		return nil
	},
}
