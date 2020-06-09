package util

import (
	"os"
	"path"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

var chartFolderEnv = os.Getenv("CHARTS_FOLDER")

func GetChartsFolder() (string, error) {
	if chartFolderEnv != "" {
		return chartFolderEnv, nil
	}

	configDir, err := GetRunaiConfigDir()

	if err != nil {
		return "", err
	}

	return path.Join(configDir, "charts"), nil
}

func StringInSlice(x string, list []string) bool {
	for _, y := range list {
		if y == x {
			return true
		}
	}
	return false
}
