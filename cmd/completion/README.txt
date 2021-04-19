
Some explanation about the completion system of the cobra package.

Further details can be found here:
    https://github.com/spf13/cobra/blob/master/shell_completions.md

The cobra completion system involves two commands, which we add to our executable:

    .-------------.  completion   .------------------.
    |             |<---script-----| runai completion |
    | shell       |               `------------------'
    | completion  |               .------------------.
    | system      |--user input-->|                  |
    |             |               | runai __complete |
    |             |<---options----|                  |
    `-------------'               `------------------'
1)
runai completion zsh|bash - outputs a completion script for the selected shell.
The help of this command provides instructions how to load the completion script, which is mainly
to source it form .zshrc/.bashrc or to add it to a /etc/... kind of directory

2)
a "hidden" command which the completion script uses, in order to get the completion options
that it sends back to the shell completion system. the command receives the words that the
user typed so far, and it returns the completion info back to the completion script.
For example:
    runai __complete submit --service-type ""
    portforward
    localhost
    nodeport
    ingress
    :4
    Completion ended with directive: ShellCompDirectiveNoFileComp
In this case the user typed "runai submit --service-type <TAB>"
And the __complete command returns the four options for service type, along with status code (NoFileComp).
Another example:
    runai __complete submit --service-type "n"
    nodeport
    :4
    Completion ended with directive: ShellCompDirectiveNoFileComp
In this case the user typed "runai submit --service-type n<TAB>", and so only nodeport option is relevant,
and this is the option that will is sent to the shell completion system.

The predictions of runai __complete are based on the definitions in cobra.Command objects. In addition to the
"usual" parameters there are also completion specific info:
    - ValidArgsFunction --> function to provide option list for args (also called nouns)
    - RegisterFlagCompletionFunction --> function to provide option list for flags

One more point to consider is that the shell completion system is responsible for the presentation of
options, and there are differences b/w the shells, with zsh being more refined.

*************
* Debugging *
*************

>> Intellij <<
Run the __complete command with the relevant parameters

>> Shell <<
export BASH_COMP_DEBUG_FILE=/tmp/<log-file>
more /tmp/<log-file>
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

**************************
*** INSTALLATION NOTES ***
**************************

Ubuntu / Debian --> Bash
------------------------
% docker run -itd ubuntu           
% docker exec -it 57c1c9281113 bash

# uname -a
Linux 57c1c9281113 4.19.121-linuxkit #1 SMP Thu Jan 21 15:36:34 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux
# bash --version
GNU bash, version 5.0.17(1)-release (x86_64-pc-linux-gnu)

# apt-get update -y
# apt-get install -y bash-completion
# source /etc/profile.d/bash_completion.sh 
# source <(runai completion bash)

# runai <TAB>
attach      completion  delete      exec        list        logout      submit      top         version     
bash        config      describe    help        login       logs        submit-mpi  update      whoami      

CentOS --> Bash
---------------
% docker run -itd centos
% docker exec -it strange_curie  bash

# uname -a
Linux fd7ab14ff907 4.19.121-linuxkit #1 SMP Thu Jan 21 15:36:34 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux
# bash --version 
GNU bash, version 4.4.19(1)-release (x86_64-redhat-linux-gnu)

# yum install bash-completion
# source /etc/profile.d/bash_completion.sh 
# source <(runai completion bash)

# runai <TAB>  
attach      completion  delete      exec        list        logout      submit      top         version     
bash        config      describe    help        login       logs        submit-mpi  update      whoami      

All -> ZSH Completion
---------------------
# apt-get install zsh  
	OR
# yum install zsh

And then:
dda6ef768e0a# autoload -U compinit; compinit -i
dda6ef768e0a# source <(runai compleiton zsh)

