package completion

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
	"os"
)

//
//   in zshell, we perform two patches the "standard" script
//   1) Adding a code to handle --flag with values
//   2) Adding a 'compdef' line in the end
//
const EXPECTED_PATCHES_ZSH = 2

//
//   this is the standard completion commands, similar to 'kubectl completion bash/zsh'
//   which the user should source into the shell in order to get completion functionality working.
//   see the README.md file for further details.
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
				response, _ := genZshCompletion(cmd)
				fmt.Print(response)
			}
		},

	}

	return cmd;
}

func genZshCompletion(cmd *cobra.Command) (string, error) {
	//
	//  we modify the zsh script slightly, so capture the default
	//  script code to a string
	//
	var script bytes.Buffer
	cmd.Root().GenZshCompletion(&script)

	//
	//   indication that we were able to modify the script properly
	//
	numPatches := 0

	result := ""

	//
	//  locate the placement of the modified code, and add the
	//  necessary lines. print all the rest as is.
	//
	for _, line := range(strings.Split(script.String(), "\n")) {
		if strings.Index(line, "compCount=0") >= 0 {
			result += `
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
`
			numPatches += 1
		}
		result += line + "\n"
	}

	//
	//   following line is needed at the end of the script, but GenZshCompetion does not
	//   print it, so add it to the output as well
	//
	result += "compdef _runai runai"
	numPatches += 1

	//
	//   check that we patched the standard script correctly. We still return the result, the patching success
	//   is checked during testing
	//
	if numPatches != EXPECTED_PATCHES_ZSH {
		return result, errors.New("Expecting " + strconv.Itoa(EXPECTED_PATCHES_ZSH) + " patches to the script, performed " + strconv.Itoa(numPatches))
	} else {
		return result, nil
	}
}

