package completion

import (
	"encoding/json"
	"github.com/run-ai/runai-cli/cmd/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

const CachingTimeSec = 5

//
//   completion function for commands with no arguments
//
func NoArgs(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func ImagePolicyValues(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string { string(v1.PullIfNotPresent), string(v1.PullNever), string(v1.PullAlways)}, cobra.ShellCompDirectiveNoFileComp
}

func ServiceTypeValues(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string { "portforward", "localhost", "nodeport", "ingress" }, cobra.ShellCompDirectiveNoFileComp
}

func OutputFormatValues(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string { "json", "yaml", "wide" }, cobra.ShellCompDirectiveNoFileComp
}

//
//   this function provides description for flag which requires some input. for example
//        submit --gpu 3
//   the standard behavior of kubectl and similar commands is not to display anything when user clicks
//   TAB in such cases. I find it more helpful to display a help message which guides the user what.
//   It will look something like this:
//	    Expecting input:	Specify number of GPUs to allocate
//   to achieve this, we'll create two "fictive" completion options: one being the text "Expecting input:" 
//   preceeded by blank (to make it appear first), and the 2nd is the help text 
//   Note that the zsh completion script is patched to avoid the completion system to add those fictive
//   options to the command being typed. in bash it is not needed
//
func AddFlagDescrpition(command *cobra.Command, name string, description string) {

	//
	//   Add backslash before any space, this is necessary for bash (without this - it considers each word
	//   of the help text as a completion option
	//
	description = strings.ReplaceAll(description, " ", "\\ ")

	command.RegisterFlagCompletionFunc(name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"\\ Expecting\\ input: ", description}, cobra.ShellCompDirectiveNoFileComp
	})
}

//
//   for parameters which take relatively long time to load (like job list or pods of job) we
//   cache the options list for 5 seconds for cases that user click TAB a few times in a raw
//   suffix is an identifier for the kind of option list that we cache
//
func ReadFromCache(suffix string) []string {

	cachePath := cacheFilePath(suffix)

	//
	//    if cache exists and relevant - use it, otherwise return nil
	//
	cacheDuration, err := util.DurationSinceLastUpdate(cachePath)
	if err != nil || cacheDuration >= (CachingTimeSec* time.Second) {
		log.Debugf("Cannot use cached value for %s", suffix)
		return nil
	}

	jsonFile, err := os.Open(cachePath)
	if err != nil {
		log.Errorf("Failed to open %s: %v", cachePath, err)
		return nil
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Errorf("Failed to read content of %s: %v", cachePath, err)
		return nil
	}

	//
	//    the list of options is cached to the file as a json array of strings
	//
	result := make([]string, 0)

	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		log.Errorf("Failed to deserialize %s: %v", cachePath, err)
		return nil
	}

	return result
}

//
//    write a list of options in a json file, to be used as a short time cache
//    for option lists which take relatively long time to build (see ReadFromCache)
//
func WriteToCache(suffix string, options []string) {

	cachePath := cacheFilePath(suffix)

	bytes, err := json.MarshalIndent(options, "", " ")
	if err != nil {
		log.Errorf("Failed to create cache %s: %v", cachePath, err)
		return
	}

	err = ioutil.WriteFile(cachePath, bytes, 0644)
	if err != nil {
		log.Errorf("Failed to write to cache %s: %v", cachePath, err)
	}
}

//
//    compose the path of a temporary JSON file containing cached array of options
//    the path will look like  /tmp/runai_<user-name>_<suffix>.json
//
func cacheFilePath(suffix string) string {

	userName := "myself"
	curUser, err := user.Current()
	if err != nil {
		log.Warnf("Failed to obtain logged in user name: %v", err)
	} else {
		userName = curUser.Username
	}

	return filepath.Join(os.TempDir(), userName + "." + suffix + ".json")
}

