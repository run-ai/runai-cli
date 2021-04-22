#Shell Completion
##Reference
https://github.com/spf13/cobra/blob/master/shell_completions.md

##Implementation Notes

### Overview

The cobra completion system involves two commands, which we add to our executable:

    .-------------.  completion   .------------------.
    |             |<---script-----| runai completion |
    | shell       |               `------------------'
    | completion  |               .------------------.
    | system      |--user input-->|                  |
    |             |               | runai __complete |
    |             |<---options----|                  |
    `-------------'               `------------------'
* runai completion zsh|bash - outputs a completion script for the selected shell.
The help of this command provides instructions how to load the completion script, which is mainly
to source it form .zshrc/.bashrc or to add it to a _/etc/..._ kind of directory.
  
* A "hidden" command which the completion script uses, in order to get the completion options
that it sends back to the shell completion system. the command receives the words that the
user typed so far, and it returns the completion info back to the completion script.
  
### Examples

The user types:
> _runai submit --service-type **TAB**_

The completion script runs:

    runai __complete submit --service-type ""

Which returns: 

    portforward
    localhost
    nodeport
    ingress
    :4
    Completion ended with directive: ShellCompDirectiveNoFileComp

And the shell displays the four options for service type.

Another example, the user types:

> _runai submit --service-type n**TAB**_

The completion scripts runs:

    runai __complete submit --service-type "n"

Which returns:

    nodeport
    :4
    Completion ended with directive: ShellCompDirectiveNoFileComp
 
This time the shell will add "nodeport" to the command, as this is the only relevant option.

### Cobra integration

The predictions of runai __complete are based on the definitions in cobra.Command objects. In addition to the
"usual" parameters there are also completion specific info:
- ValidArgsFunction --> function to provide option list for args (also called nouns)
- RegisterFlagCompletionFunction --> function to provide option list for flags

One more point to consider is that the shell completion system is responsible for the presentation of
options, and there are differences b/w the shells, with zsh being more refined.

## Debugging

### Intellij 

Run the __complete command with the relevant parameters

### Shell 

    $ export BASH_COMP_DEBUG_FILE=/tmp/<log-file>

Type runai command, and click TABs where you want the shell to call the completion script.

To check the log:

    $ more /tmp/<log-file>
    ========= starting completion logic ==========
    CURRENT: 4, words[*]: runai describe node
    Truncated words[*]: runai describe node ,
    lastParam: , lastChar:
    Adding extra empty parameter
    About to call: eval runai __complete describe node  ""
    completion output: dev-ofer-master
    dev-ofer-worker-cpu
    dev-ofer-worker-gpu-1
    :4
    last line: :4
    directive: 4
    completions: dev-ofer-master
    dev-ofer-worker-cpu
    dev-ofer-worker-gpu-1
    flagPrefix:
    Adding completion: dev-ofer-master
    Adding completion: dev-ofer-worker-cpu
    Adding completion: dev-ofer-worker-gpu-1

## Installation Notes

### Ubuntu / Debian (Bash)

Tested on the following docker:

    % docker run -itd ubuntu           
    % docker exec -it 57c1c9281113 bash
    # uname -a
    Linux 57c1c9281113 4.19.121-linuxkit #1 SMP Thu Jan 21 15:36:34 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux
    # bash --version
    GNU bash, version 5.0.17(1)-release (x86_64-pc-linux-gnu)

Installation steps:

    apt-get update -y
    apt-get install -y bash-completion

Adding to .bashrc:

    source /etc/profile.d/bash_completion.sh 
    source <(runai completion bash)

After restarting the shell, test that it works:

    # runai <TAB>
    attach      completion  delete      exec        list        logout      submit      top         version     
    bash        config      describe    help        login       logs        submit-mpi  update      whoami      

## CentOS (Bash)

Tested on the following docker:

    % docker run -itd centos
    % docker exec -it strange_curie  bash
    # uname -a
    Linux fd7ab14ff907 4.19.121-linuxkit #1 SMP Thu Jan 21 15:36:34 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux
    # bash --version 
    GNU bash, version 4.4.19(1)-release (x86_64-redhat-linux-gnu)

Installation steps:

    yum install bash-completion

Add to .bashrc:

    # source /etc/profile.d/bash_completion.sh 
    # source <(runai completion bash)

After restarting the shell, test that it works:

    # runai <TAB>  
    attach      completion  delete      exec        list        logout      submit      top         version     
    bash        config      describe    help        login       logs        submit-mpi  update      whoami      

## Debian/Centos/Ubunto -> ZSH

Installation steps:

    # apt-get install zsh

Or: 

    # yum install zsh

Add to .zshrc:

    autoload -U compinit; compinit -i
    source <(runai compleiton zsh)

