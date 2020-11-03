package flags

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai-cli/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagGroupName string

type FlagGroup struct {
	Name    FlagGroupName
	FlagSet *pflag.FlagSet
}

type FlagsByGroups struct {
	cmd           *cobra.Command
	groupsByOrder *[]FlagGroup
}

func NewFlagsByGroups(cmd *cobra.Command) FlagsByGroups {
	return FlagsByGroups{
		cmd:           cmd,
		groupsByOrder: &([]FlagGroup{}),
	}
}

func (fg *FlagGroup) usage() string {
	return fmt.Sprint(ui.Bold(fg.Name), "\n", fg.FlagSet.FlagUsagesWrapped(1))
}

func NewFlagGroup(name FlagGroupName) FlagGroup {
	return FlagGroup{
		Name:    name,
		FlagSet: pflag.NewFlagSet(string(name), 0),
	}
}

func (fbg *FlagsByGroups) GetOrAddFlagSet(groupName FlagGroupName) (fs *pflag.FlagSet) {
	for _, item := range *fbg.groupsByOrder {
		if item.Name == groupName {
			fs = item.FlagSet
		}
	}
	if fs == nil {
		newFlagGroup := NewFlagGroup(groupName)
		*fbg.groupsByOrder = append(*fbg.groupsByOrder, newFlagGroup)
		fs = newFlagGroup.FlagSet
	}
	return fs
}

func (fbg *FlagsByGroups) UpdateFlagsByGroupsToCmd() {
	updateFlagsByGroupsToCmd(fbg.cmd, *fbg.groupsByOrder...)
}

func updateFlagsByGroupsToCmd(cmd *cobra.Command, fgs ...FlagGroup) {
	for _, fg := range fgs {
		cmd.Flags().AddFlagSet(fg.FlagSet)
	}
	usage := flagsGroupUsage(fgs...)

	cmd.SetUsageTemplate(
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
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}

func flagsGroupUsage(fgs ...FlagGroup) string {
	usage := []string{}
	for _, fg := range fgs {
		usage = append(usage, fg.usage())
	}

	return strings.Join(usage, "\n\n")
}
