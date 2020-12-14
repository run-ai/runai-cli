package submit

import (
	"gotest.tools/assert"
	"testing"
)

func TestGitSyncConnectionStringFullCS(t *testing.T) {
	connectionString := "source=repo-url,rev=tag/git-hash,branch=branch-name,target=path/to/sync/to,username=username,password=asd"

	gitSync := GitSyncFromConnectionString(connectionString)

	assert.Equal(t, gitSync.Repository, "repo-url")
	assert.Equal(t, gitSync.Revision, "tag/git-hash")
	assert.Equal(t, gitSync.Branch, "branch-name")
	assert.Equal(t, gitSync.Directory, "path/to/sync/to")
	assert.Equal(t, gitSync.Username, "username")
	assert.Equal(t, gitSync.Password, "asd")
}

func TestGitSyncConnectionStringPartialCS(t *testing.T) {
	connectionString := "source=repo-url,branch=branch-name,target=path/to/sync/to"

	gitSync := GitSyncFromConnectionString(connectionString)

	assert.Equal(t, gitSync.Repository, "repo-url")
	assert.Equal(t, gitSync.Branch, "branch-name")
	assert.Equal(t, gitSync.Directory, "path/to/sync/to")
}

func TestGitSyncConnectionStringContainParentheses(t *testing.T) {
	connectionString := "source=repo-url,branch=\"branch-name\",target='path/to/sync/to'"

	gitSync := GitSyncFromConnectionString(connectionString)

	assert.Equal(t, gitSync.Repository, "repo-url")
	assert.Equal(t, gitSync.Branch, "branch-name")
	assert.Equal(t, gitSync.Directory, "path/to/sync/to")
}
