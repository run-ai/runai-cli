package util

import (
	"os"
	"path"
	"path/filepath"
)

var (
	configDir = ""
)

func AddNamespaceToArgs(args []string, namespace string) []string {
	if namespace == "" {
		return args
	}

	return append(args, "--namespace", namespace)
}

func GetRunaiConfigDir() (string, error) {
	if configDir != "" {
		return configDir, nil
	}

	dir, err := os.Executable()
	if err != nil {
		return "", err
	}

	realPath, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}

	return path.Dir(realPath), nil
}
