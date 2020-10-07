package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)


type FlagsGroup struct {
	Title string
	FlagSet * pflag.FlagSet
}

type MapFlagsGroup struct {
	cmd *cobra.Command
	m map[string]FlagsGroup
	order *[]string
}

func NewMapFlagsGroup(cmd *cobra.Command) MapFlagsGroup {
	return MapFlagsGroup {
		cmd: cmd,
		m: map[string]FlagsGroup{},
		order: &([]string{}),
	}
}

func (fg *FlagsGroup) Usage() string {
	return fmt.Sprint( fg.Title, "\n", fg.FlagSet.FlagUsagesWrapped(1))
}

func NewFlagsGroup(title string) FlagsGroup {
	return FlagsGroup {
		Title: title,
		FlagSet: pflag.NewFlagSet(title, 0),
	}
}

func (mfg MapFlagsGroup) GetOrAddFlagSet(groupName string) *pflag.FlagSet {
	fg, found := mfg.m[groupName]
	if !found {
		fg = NewFlagsGroup(groupName)
		mfg.m[groupName] = fg
		*mfg.order = append(*mfg.order, groupName)
	}
	return fg.FlagSet
}

func (mfg MapFlagsGroup) Groups() []FlagsGroup {
	groups := []FlagsGroup{}
	for _, name := range *mfg.order {
		groups = append(groups, mfg.m[name])
	}
	return groups
}

func (mfg MapFlagsGroup) ConnectToCmd() {
	ConnectFlagsGroupToCmd(mfg.cmd, mfg.Groups()...)
}


// ConnectFlagsGroupToCmd
func ConnectFlagsGroupToCmd(cmd *cobra.Command, fgs ...FlagsGroup) {
	for _, fg := range fgs {
		cmd.Flags().AddFlagSet(fg.FlagSet)
	}
	usage := FlagsGroupUsage(fgs...)

	cmd.SetUsageTemplate(fmt.Sprintf(
`Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:

`+usage+`{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}`))
}

func FlagsGroupUsage(fgs ...FlagsGroup) string {
	usage := []string{}
	for _, fg := range fgs {
		usage = append(usage, fg.Usage())
	}

	return strings.Join(usage,"\n\n")
}