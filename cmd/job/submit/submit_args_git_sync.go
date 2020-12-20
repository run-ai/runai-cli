package submit

import (
	"fmt"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"strings"
)

const (
	defaultGitSyncImage  = "k8s.gcr.io/git-sync/git-sync:v3.2.0"
	defaultSyncDirectory = "/code"
	defaultBranch        = "master"
)

type GitSync struct {
	Sync           *bool  `yaml:"sync,omitempty"`
	Image          string `yaml:"image,omitempty"`
	Repository     string `yaml:"repository,omitempty"`
	ByRevision     bool   `yaml:"byRevision,omitempty"`
	Branch         string `yaml:"branch,omitempty"`
	Revision       string `yaml:"revision,omitempty"`
	Username       string `yaml:"username,omitempty"`
	Password       string `yaml:"password,omitempty"`
	UseCredentials bool   `yaml:"useCredentials,omitempty"`
	Directory      string `yaml:"directory,omitempty"`
}

func NewGitSync() *GitSync {
	notSync := false
	return &GitSync{
		Sync:       &notSync,
		Image:      "",
		ByRevision: false,
		Branch:     "",
		Revision:   "",
		Username:   "",
		Password:   "",
		Directory:  "",
	}
}

func (gs *GitSync) HandleGitSync() error {
	if !raUtil.IsBoolPTrue(gs.Sync) {
		return nil
	}

	if gs.Revision != "" {
		gs.ByRevision = true
	} else if gs.Branch == "" {
		gs.Branch = defaultBranch
	}
	if gs.Username != "" || gs.Password != "" {
		gs.UseCredentials = true
	}
	if gs.Image == "" {
		gs.Image = defaultGitSyncImage
	}
	if gs.Directory == "" {
		gs.Directory = defaultSyncDirectory
	}

	if gs.Repository == "" {
		return fmt.Errorf("git sync must contain Repository")
	}
	return nil
}

func GitSyncFromConnectionString(connectionString string) *GitSync {
	parameters := strings.Split(connectionString, ",")

	arguments := make(map[string]string)
	for _, parameter := range parameters {
		keyValuePair := strings.Split(parameter, "=")
		if len(keyValuePair) != 2 {
			continue
		}
		value := strings.Trim(keyValuePair[1], "'\"")
		arguments[keyValuePair[0]] = value
	}

	if len(arguments) == 0 {
		return nil
	}

	return ParseGitSyncArguments(arguments)
}

func ParseGitSyncArguments(arguments map[string]string) *GitSync {
	syncObject := NewGitSync()
	tv := true
	syncObject.Sync = &tv
	for key, value := range arguments {
		switch key {
		case "source":
			syncObject.Repository = value
		case "branch":
			syncObject.Branch = value
		case "rev":
			syncObject.Revision = value
		case "image":
			syncObject.Image = value
		case "username":
			syncObject.Username = value
		case "password":
			syncObject.Password = value
		case "target":
			syncObject.Directory = value
		}
	}
	return syncObject
}
