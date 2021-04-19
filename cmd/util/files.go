package util

import (
	"os"
	"time"
)

//
//   calculate the duration since the last time a file has been modified
//
func DurationSinceLastUpdate(path string) (time.Duration, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	duration := time.Time.Sub(time.Now(), info.ModTime())
	return duration, nil
}
