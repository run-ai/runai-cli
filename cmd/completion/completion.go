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

  # bash completion is assumed to be enabled
  $ source <(runai completion bash)

Zsh:

  $ autoload -U compinit; compinit -i
  $ source <(runai completion zsh) 
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
