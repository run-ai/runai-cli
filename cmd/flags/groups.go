package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagGroup struct {
	Title   string
	FlagSet *pflag.FlagSet
}

type FlagGroupMap struct {
	cmd   *cobra.Command
	m     map[string]FlagGroup
	order *[]string
}

func NewFlagGroupMap(cmd *cobra.Command) FlagGroupMap {
	return FlagGroupMap{
		cmd:   cmd,
		m:     map[string]FlagGroup{},
		order: &([]string{}),
	}
}

func (fg *FlagGroup) Usage() string {
	return fmt.Sprint(fg.Title, "\n", fg.FlagSet.FlagUsagesWrapped(1))
}

func NewFlagsGroup(title string) FlagGroup {
	return FlagGroup{
		Title:   title,
		FlagSet: pflag.NewFlagSet(title, 0),
	}
}

func (fgm FlagGroupMap) GetOrAddFlagSet(groupName string) *pflag.FlagSet {
	fg, found := fgm.m[groupName]
	if !found {
		fg = NewFlagsGroup(groupName)
		fgm.m[groupName] = fg
		*fgm.order = append(*fgm.order, groupName)
	}
	return fg.FlagSet
}

func (fgm FlagGroupMap) Groups() []FlagGroup {
	groups := []FlagGroup{}
	for _, name := range *fgm.order {
		groups = append(groups, fgm.m[name])
	}
	return groups
}

func (fgm FlagGroupMap) ConnectToCmd() {
	ConnectFlagsGroupToCmd(fgm.cmd, fgm.Groups()...)
}

// ConnectFlagsGroupToCmd
func ConnectFlagsGroupToCmd(cmd *cobra.Command, fgs ...FlagGroup) {
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
` + usage + `{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}`))
}

func FlagsGroupUsage(fgs ...FlagGroup) string {
	usage := []string{}
	for _, fg := range fgs {
		usage = append(usage, fg.Usage())
	}

	return strings.Join(usage, "\n\n")
}
