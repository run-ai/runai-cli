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
	return fmt.Sprint(fg.Title, ":\n", fg.FlagSet.FlagUsagesWrapped(1))
}

func NewFlagGroup(title string) FlagGroup {
	return FlagGroup{
		Title:   title,
		FlagSet: pflag.NewFlagSet(title, 0),
	}
}

func (fbg *FlagsByGroups) GetOrAddFlagSet(groupName string) (fs *pflag.FlagSet){
	for _, item := range *fbg.groupsByOrder {
		if item.Title == groupName {
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
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}`)
}

func flagsGroupUsage(fgs ...FlagGroup) string {
	usage := []string{}
	for _, fg := range fgs {
		usage = append(usage, fg.usage())
	}

	return strings.Join(usage, "\n\n")
}
