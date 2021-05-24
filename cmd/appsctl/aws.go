package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var shouldCreate bool
var shouldCreateAccessKey bool
var userName string
var policyName string
var groupName string

func init() {
	rootCmd.AddCommand(
		awsCmd,
	)

	// init
	awsCmd.AddCommand(awsInitCmd)
	awsInitCmd.Flags().BoolVar(&shouldCreate, "create", false, "Create resources (user, group, policy, bucket) that don't already exist, using the default configuration.")
	awsInitCmd.Flags().BoolVar(&shouldCreateAccessKey, "create-access-key", false, "Create new access key for the user (or you can safely do it in AWS Console).")
	awsInitCmd.Flags().StringVar(&userName, "user", upaws.DefaultUserName, "Username to use for invoking the AWS App from Mattermost Server.")
	awsInitCmd.Flags().StringVar(&policyName, "policy", upaws.DefaultPolicyName, "Name of the policy to control access of AWS services directly by Mattermost Server (user).")
	awsInitCmd.Flags().StringVar(&groupName, "group", upaws.DefaultGroupName, "Name of the user group connecting the invoking user to the invoke policy.")

	// clean
	awsCmd.AddCommand(awsCleanCmd)

	// test
	awsCmd.AddCommand(awsTestCmd)
	awsTestCmd.AddCommand(awsTestLambdaCmd)
	awsTestCmd.AddCommand(awsTestProvisionCmd)
	awsTestCmd.AddCommand(awsTestS3Cmd)
}

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Manage AWS upstream for Mattermost Apps",
}

var awsInitCmd = &cobra.Command{
	Use:   "aws",
	Short: "Initialize AWS to deploy Mattermost Apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		asProvisioner, err := AsProvisioner()
		if err != nil {
			return err
		}

		out, err := upaws.InitializeAWS(asProvisioner, &log, upaws.InitParams{
			Bucket:                upaws.S3BucketName(),
			User:                  upaws.Name(userName),
			Group:                 upaws.Name(groupName),
			Policy:                upaws.Name(policyName),
			ExecuteRole:           upaws.Name(executeRoleName),
			ShouldCreate:          shouldCreate,
			ShouldCreateAccessKey: shouldCreateAccessKey,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Ready to deploy AWS Lambda Mattermost Apps!\n\n")

		fmt.Printf("User:\t%q\n", out.UserARN)
		fmt.Printf("Group:\t%q\n", out.GroupARN)
		fmt.Printf("Policy:\t%q\n", out.PolicyARN)
		fmt.Printf("Bucket:\t%q\n", out.Bucket)

		if shouldCreateAccessKey {
			fmt.Printf("\nPlease store the Access Key securely, it will not be viewable again.\n\n")
			fmt.Printf("export %s='%s'\n", upaws.AccessEnvVar, out.AccessKeyID)
			fmt.Printf("export %s='%s'\n", upaws.SecretEnvVar, out.AccessKeySecret)
		}

		return nil
	},
}

var awsCleanCmd = &cobra.Command{
	Use:   "aws",
	Short: "Initialize AWS to deploy Mattermost Apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		asProvisioner, err := AsProvisioner()
		if err != nil {
			return err
		}

		accessKeyID := os.Getenv(upaws.AccessEnvVar)
		if accessKeyID == "" {
			return errors.Errorf("no AWS access key was provided. Please set %s", upaws.AccessEnvVar)
		}

		return upaws.CleanAWS(asProvisioner, accessKeyID, &log)
	},
}

var awsTestCmd = &cobra.Command{
	Use:   "test",
	Short: "test accessing a provisioned resource",
}

var helloApp = &apps.App{
	Manifest: apps.Manifest{
		AppID:   "hello-lambda",
		Version: "demo",
	},
}

var awsTestS3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "test accessing a static S3 resource",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		asTest, err := AsTest()
		if err != nil {
			return err
		}

		up := upaws.NewUpstream(helloApp, asTest, upaws.S3BucketName())
		resp, _, err := up.GetStatic("test.txt")
		if err != nil {
			return err
		}
		defer resp.Close()
		data, err := io.ReadAll(resp)
		if err != nil {
			return err
		}
		r := string(data)
		log.Debugf("received: %s", string(data))

		if r != "static pong" {
			return errors.Errorf("expected 'static pong', got '%s'", r)
		}
		fmt.Println("OK")
		return nil
	},
}

var awsTestLambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "test accessing hello-lambda /ping function",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		asTest, err := AsTest()
		if err != nil {
			return err
		}

		app := &apps.App{
			Manifest: apps.Manifest{
				AppID: "hello-lambda",
			},
		}
		up := upaws.NewUpstream(app, asTest, "")
		crString, err := up.InvokeFunction("hello-lambda_demo_go-function", false, &apps.CallRequest{
			Call: apps.Call{
				Path: "/ping",
			},
		})
		if err != nil {
			return err
		}
		log.Debugf("received: %s", crString)

		cr := apps.CallResponse{}
		_ = json.Unmarshal([]byte(crString), &cr)
		expected := apps.CallResponse{Markdown: "PONG", Type: apps.CallResponseTypeOK}
		if cr != expected {
			return errors.Errorf("invalid value received: %s", crString)
		}

		fmt.Println("OK")
		return nil
	},
}

var awsTestProvisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "provisions 'hello-lambda' app for testing",
	Long: `Test commands us the 'hello-lambda' example app for testing, see
https://github.com/mattermost/mattermost-plugin-apps/tree/master/examples/go/hello-lambda/README.md

The App needs to be built with 'make dist' in its own directory, then use

	appsctl aws test provision <dist-bundle-path>

command to provision it to AWS. This command is equivalent to

	appsctl provision app <dist-bundle-path> --update

with the default initial IAM configuration`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]

		asProvisioner, err := AsProvisioner()
		if err != nil {
			return err
		}

		out, err := upaws.ProvisionAppFromFile(asProvisioner, bundlePath, &log, upaws.ProvisionAppParams{
			Bucket:           upaws.S3BucketName(),
			InvokePolicyName: upaws.Name(upaws.DefaultPolicyName),
			ExecuteRoleName:  upaws.Name(upaws.DefaultExecuteRoleName),
			ShouldUpdate:     true,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Success!\n\n%s\n", utils.Pretty(out))
		return nil
	},
}
