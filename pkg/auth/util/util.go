package util

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"strings"
	"syscall"
)

// ReadString reads a string from the stdin.
func ReadString(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	r := bufio.NewReader(os.Stdin)
	s, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	s = strings.TrimRight(s, "\r\n")
	return s, nil
}

// ReadPassword reads a password from the stdin without echo back.
func ReadPassword(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	b, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	if _, err := fmt.Fprintln(os.Stderr); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	return string(b), nil
}

func MakeNonce() string {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		log.Debug(err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buffer)
}

func ReadKubeConfig() (*clientapi.Config, error) {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return configAccess.GetStartingConfig()
}

func WriteKubeConfig(config *clientapi.Config) error {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return clientcmd.ModifyConfig(configAccess, *config, true)
}

func MergeScopes(defaultScopes []string, extraScopes []string) (scopes []string) {
	set := make(map[string]struct{})
	for _, scope := range defaultScopes {
		set[scope] = struct{}{}
	}
	for _, scope := range extraScopes {
		set[scope] = struct{}{}
	}
	for scope, _ := range set {
		scopes = append(scopes, scope)
	}
	return
}