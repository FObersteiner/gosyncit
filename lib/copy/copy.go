package copy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gosyncit/lib/config"
	"gosyncit/lib/fileset"
)

const (
	BUFFERSIZE                  = 4096
	defaultModeDir  fs.FileMode = 0755 // rwxr-xr-x
	defaultModeFile fs.FileMode = 0644 // rw-r--r--
)

type CopyFunc func(string, string, *config.Config) error

// SelectCopyFunc selects a copy function based on the status of 'dst'.
// If 'dst' is empty or does not exist, a simple mirror without cleaning suffices.
// If 'dst' exists and is not empty, mirror or sync are appropriate.
func SelectCopyFunc(dst string, c *config.Config) CopyFunc {
	isEmpty, err := isEmptyDir(dst)
	if isEmpty || err != nil {
		return MirrorBasic
	}
	if c.Sync {
		return Sync
	}
	return Mirror
}

// MirrorBasic performs a one-way copy from src to dst
func MirrorBasic(src, dst string, c *config.Config) error {
	return filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if c.IgnoreHidden && strings.Contains(path, "/.") {
				c.Log.Info().Msgf("skip   : hidden file '%s'", path)
				return nil
			}
			childPath := strings.TrimPrefix(path, src)
			dstPath := filepath.Join(dst, childPath)
			if !srcInfo.IsDir() && !srcInfo.Mode().IsRegular() {
				// TODO : this might be the place to handle symlinks
				c.Log.Info().Msgf("skip non-regular file '%v'", path)
				return nil
			}
			c.Log.Info().Msgf("copy / create dir '%v'", dstPath)
			return copyFileOrCreateDir(path, dstPath, srcInfo, c)
		},
	)
}

// Mirror copies newer files to dst and removes files from dst that do not
// exist in src
func Mirror(src, dst string, c *config.Config) error {
	c.Log.Info().Msg("MIRROR")
	filesetSrc, err := fileset.New(src)
	if err != nil {
		return err
	}
	filesetDst, err := fileset.New(dst)
	if err != nil {
		return err
	}
	// make filesets concurrently, sync output via error channel
	errChan := make(chan error)
	go func() {
		errChan <- filesetSrc.Populate()
	}()
	go func() {
		errChan <- filesetDst.Populate()
	}()
	err = <-errChan
	if err != nil {
		<-errChan // make sure other goroutine has completed TODO : needed ?!
		return err
	}
	err = <-errChan
	if err != nil {
		return err
	}
	c.Log.Info().Msg(filesetSrc.String())
	c.Log.Info().Msg(filesetDst.String())
	// step 1: copy everything from source to dst if src newer
	// for file in filesetSrc: src file exists in dst ?
	//   yes --> compare, copy if newer
	//   no --> copy
	//
	// step 2: clean everything from dst that is not in src
	// -- if c.CleanDst:
	// for file in filesetDst: file exists in filesetSrc ?
	//   yes --> continue
	//   no --> delete
	return nil
}

// Sync synchronizes files between src and dst, keeping only the newer versions
func Sync(src, dst string, c *config.Config) error {
	c.Log.Info().Msg("SYNC")
	// step 1: copy everything from source to dst if src newer or file does not exist dst
	// step 2: copy everything from dst to src if dst newer or file does not exist in src
	return nil
}

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

// ------------------------------------------------------------------------------------------------

func isEmptyDir(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func copyFileOrCreateDir(src, dst string, sourceFileStat fs.FileInfo, c *config.Config) error {
	if c.DryRun {
		return nil
	}
	if sourceFileStat.IsDir() {
		return os.MkdirAll(dst, defaultModeDir)
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	err := CopyFile(src, dst)
	if err != nil {
		return err
	}
	if c.KeepPermissions && runtime.GOOS != "windows" {
		return os.Chmod(dst, sourceFileStat.Mode())
	}
	return nil
}
