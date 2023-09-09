package copy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gosyncit/lib/pathlib"
)

const (
	BUFFERSIZE                  = 4096
	defaultModeDir  fs.FileMode = 0755 // rwxr-xr-x
	defaultModeFile fs.FileMode = 0644 // rw-r--r--
)

// SimpleCopy makes a simple copy of directory 'src' to directory 'dst'.
func SimpleCopy(src, dst string, dry bool) error {
	src, dst, err := pathlib.CheckSrcDst(src, dst)
	if err != nil {
		return err
	}

	return filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			childPath := strings.TrimPrefix(path, src)
			dstPath := filepath.Join(dst, childPath)
			if !srcInfo.IsDir() && !srcInfo.Mode().IsRegular() {
				fmt.Printf("skip non-regular file '%v'\n", path)
				return nil
			}
			fmt.Printf("copy/create file/dir '%v'\n", dstPath)
			return copyFileOrCreateDir(path, dstPath, srcInfo, dry)
		},
	)
}

// copyFileOrCreateDir is wraps the creation of a directory or a file, depending
// on the src type.
func copyFileOrCreateDir(src, dst string, sourceFileStat fs.FileInfo, dry bool) error {
	if dry {
		return nil
	}
	if sourceFileStat.IsDir() {
		return os.MkdirAll(dst, defaultModeDir)
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	return CopyFile(src, dst)
}

// CopyFile copies src to dst. If dst exists, it will be overwritten.
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

// CopyPerm tries to copy permissions from src to dst file
func CopyPerm(src, dst string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcStat.Mode())
}
