package submit

import (
	"fmt"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
)

const (
	defaultGitSyncImage  = "k8s.gcr.io/git-sync/git-sync:v3.2.0"
	defaultSyncDirectory = "/code"
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
	return &GitSync{
		Image:      "",
		ByRevision: false,
		Branch:     "master",
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

	if gs.Repository == "" || (gs.Branch == "" && gs.Revision == "") {
		return fmt.Errorf("git sync must contain Repository, and branch or revision")
	}
	return nil
}
