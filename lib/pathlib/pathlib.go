package pathlib

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	path = Resolve(filepath.Clean(path))
	stats, err := os.Stat(path)
	if os.IsNotExist(err) {
		return path, fmt.Errorf("specified directory '%v' does not exist", path)
	}
	if err != nil { // some other error ?!
		return path, err
	}
	if !stats.IsDir() {
		return path, fmt.Errorf("specified directory '%v' is a file", path)
	}
	return path, nil
}

// ResolveHomeDir converts potential home dir ~ into an absolute path
func ResolveHomeDir(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}
	return path
}

// Resolve converts potential home dir ~ and cd .. into an absolute path
func Resolve(path string) string {
	path = ResolveHomeDir(path)
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}
