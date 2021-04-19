package completion

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

//
//   this is the standard completion commands, similar to 'kubectl completion bash/zsh'
//   which the user should source into the shell in order to get completion functionality working.
//   see the README.txt file for further details.
//   the output of this command mainly explain to the user how to load the compleiton definitions
//   into his shell (slightly different b/w the various unix shells).
//
func NewCompletionCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "completion [bash|zsh]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(runai completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ runai completion bash > /etc/bash_completion.d/runai
  # macOS:
  $ runai completion bash > /usr/local/etc/bash_completion.d/runai

Zsh:

  $ source <(runai completion zsh) 

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ runai completion zsh > "${fpath[1]}/_runai"

  # You will need to start a new shell for this setup to take effect.
`,

		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh" },
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
				//   following line is needed, but GenZshCompetion does not print it
				fmt.Print("compdef _runai runai")
			}
		},
	}

	return cmd;
}
