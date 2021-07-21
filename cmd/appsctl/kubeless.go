package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
)

// var (
// 	shouldCreate          bool
// 	shouldCreateAccessKey bool
// 	userName              string
// 	policyName            string
// 	groupName             string
// 	shouldUpdate          bool
// 	invokePolicyName      string
// 	executeRoleName       string
// )

func init() {
	rootCmd.AddCommand(kubelessCmd)

	// provision
	kubelessCmd.AddCommand(kubelessProvisionCmd)
	kubelessProvisionCmd.Flags().BoolVar(&shouldUpdate, "update", false, "Update functions if they already exist. Use with causion in production.")

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

		m, err := upkubeless.ProvisionApp(bundlePath, &log, shouldUpdate)
		if err != nil {
			return err
		}
		if len(m.KubelessFunctions) == 0 {
			return errors.New("\nno functions to provision, check manifest.json\n")
		}

		fmt.Printf("\nProvisioned '%s' to Kubeless, %v functions deployed.\n", m.DisplayName, len(m.KubelessFunctions))
		fmt.Printf("You can now install it in Mattermost using:\n")
		fmt.Printf("  /apps install kubeless <manifest URL>\n\n")
		return nil
	},
}

var kubelessTestCmd = &cobra.Command{
	Use:   "test",
	Short: "provisions and tests 'hello-lambda'",
	Long: `Test commands us the 'hello-lambda' example app for testing, see
https://github.com/mattermost/mattermost-plugin-apps/tree/master/examples/go/hello-lambda/README.md

The App needs to be built with 'make dist' in its own directory, then use
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// bundlePath := args[0]
		return nil
	},
}

func makeTestKubelessUpstream() (*upkubeless.Upstream, error) {
	return upkubeless.MakeUpstream()
}
