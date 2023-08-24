package pathlib

import (
	"errors"
	"os"
	"path/filepath"
)

func CheckPath(path string) (string, error) {
	path = filepath.Clean(path)
	stats, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !stats.IsDir() {
		return "", errors.New("specified directory is a file")
	}
	return path, nil
}
