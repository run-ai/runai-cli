package auth

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth/jwt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"sort"
)

const CacheLocationSuffix = "/.kube/cache/oidc-login"

func GetKubeLoginErrorIfNeeded(err error) error {
	if isAuthError(err) {
		err = getKubeLoginError(err)
	}
	return err
}

func getKubeLoginError(err error) error {
	if username, kubeloginErr := GetEmailForCurrentKubeloginToken(); kubeloginErr != nil {
		log.Debug("Can't acquire username from kubelogin token cache: ", kubeloginErr)
	} else if username != "" {
		//Write the original message to debug log so we can actually understand what's going on.
		log.Debug(err)
		err = fmt.Errorf("user %s doesn't have the required permissions to perform this operation", username)
	}
	return err
}

func isAuthError(err error) bool {
	return errors.IsForbidden(err) || errors.IsUnauthorized(err)
}

// This is a best effort to:
// 1. Locate the newest token file in kubelogin's default cache location (in case there are several)
// 2. Parse the token
// 3. Match the username string that appears in the error to the user's email (which is how that user is represented in the UI).
// If all of the above apply the user's email will be returned
func GetEmailForCurrentKubeloginToken() (email string, err error) {
	var tokenFilePath string
	var token jwt.Token
	if tokenFilePath, err = getNewestTokenFile(); err == nil {
		if token, err = jwt.DecodeTokenFile(tokenFilePath); err == nil {
			email = token.Email
		}
	}
	return email, err
}

func getNewestTokenFile() (tokenFilePath string, err error) {
	var homeDir string
	var tokenFiles []os.FileInfo
	if homeDir, err = os.UserHomeDir(); err == nil {
		kubeloginCacheDir := homeDir + CacheLocationSuffix
		if tokenFiles, err = ioutil.ReadDir(kubeloginCacheDir); err == nil && len(tokenFiles) > 0 {
			sort.Sort(ByModTime(tokenFiles))
			tokenFilePath = kubeloginCacheDir + "/" + tokenFiles[0].Name()
		}
	}
	return tokenFilePath, err
}

// For easy sorting of a file list based on modification time
type ByModTime []os.FileInfo

func (files ByModTime) Len() int {
	return len(files)
}

func (files ByModTime) Swap(i, j int) {
	files[i], files[j] = files[j], files[i]
}

func (files ByModTime) Less(i, j int) bool {
	return files[i].ModTime().Before(files[j].ModTime())
}
