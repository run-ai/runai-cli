
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

