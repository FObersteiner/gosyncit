package pathlib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CheckSrcDst cleans input path strings 'src' and 'dst'. Src and dst must not be identical,
// src must exist but dst not necessarily.
func CheckSrcDst(src, dst string) (string, string, error) {
	srcCleaned, err := CheckDirPath(src)
	if err != nil {
		return "", "", err
	}
	dstCleaned, err := CheckDirPath(dst)
	// If dst directory does not exist, that is not considered an error.
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}
	if src == dst {
		return "", "", errors.New("src and dst must not be identical")
	}
	return srcCleaned, dstCleaned, nil
}

// CheckDirPath cleans a path and checks for its existence.
// Files are considered invalid, must be directory.
// Non-existing directories are considered an error here.
func CheckDirPath(path string) (string, error) {
	path = filepath.Clean(path)
	stats, err := os.Stat(path)
	if err != nil {
		return path, err
	}
	if !stats.IsDir() {
		return path, fmt.Errorf("specified '%v' directory is a file", path)
	}
	return path, nil
}
