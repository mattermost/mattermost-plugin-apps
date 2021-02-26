package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(pluginops completion bash)

# To load completions for each session, execute once:
Linux:
  $ pluginops completion bash > /etc/bash_completion.d/pluginops
MacOS:
  $ pluginops completion bash > /usr/local/etc/bash_completion.d/pluginops

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ pluginops completion zsh > "${fpath[1]}/_pluginops"

# You will need to start a new shell for this setup to take effect.

Fish:

$ pluginops completion fish | source

# To load completions for each session, execute once:
$ pluginops completion fish > ~/.config/fish/completions/pluginops.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletion(os.Stdout)
		default:
			return errors.Errorf("unknown sheell type %v", args[0])
		}
	},
}
