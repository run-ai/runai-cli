package completion

import (
	"bytes"
	"fmt"
	"strings"
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
				//
				//  we modify the zsh script slightly, so capture the default
				//  script code to a string
				//
				var script bytes.Buffer
				cmd.Root().GenZshCompletion(&script)

				//
				//  locate the placement of the modified code, and add the
				//  necessary lines. print all the rest as is.
				//
				for _, line := range(strings.Split(script.String(), "\n")) {
					if strings.Index(line, "compCount=0") >= 0 {
						fmt.Println(`
    #======================================================
    # Special treatment for displaying flags description
    #======================================================	
    if [[ "$out" == *Expecting*input* ]] ; then
        zstyle ':completion:*:runai:*' format $(echo $out | sed 's/\\//g')
        _message
        return
    else
        zstyle ':completion:*:runai:*' format ""
    fi
`)
					}
					fmt.Println(line)
				}
				//
				//   following line is needed at the end of the script, but GenZshCompetion does not
				//   print it, so add it to the output as well
				//
				fmt.Println("compdef _runai runai")
			}
		},
	}

	return cmd;
}
