package copy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gosyncit/lib/compare"
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

// MirrorBasic performs a one-way copy from src to dst.
// Any content in dst will be ignored (overwritten)
func MirrorBasic(src, dst string, c *config.Config) error {
	c.Log.Debug().Msg("MIRROR BASIC")
	return filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if c.IgnoreHidden && strings.Contains(path, "/.") {
				c.Log.Debug().Msgf("skip   : hidden file '%s'", path)
				return nil
			}
			childPath := strings.TrimPrefix(path, src)
			dstPath := filepath.Join(dst, childPath)
			if !srcInfo.IsDir() && !srcInfo.Mode().IsRegular() {
				c.Log.Debug().Msgf("skip non-regular file '%v'", path)
				return nil
			}
			c.Log.Debug().Msgf("copy/create file/dir '%v'", dstPath)
			return copyFileOrCreateDir(path, dstPath, srcInfo, c)
		},
	)
}

// Mirror copies newer files to dst and removes files from dst that do not
// exist in src
func Mirror(src, dst string, c *config.Config) error {
	c.Log.Debug().Msg("MIRROR")
	filesetSrc, err := fileset.New(src)
	if err != nil {
		return err
	}
	filesetDst, err := fileset.New(dst)
	if err != nil {
		return err
	}

	// we need a fileset for the destination, to check against while walking the src
	err = filesetDst.Populate()
	if err != nil {
		return err
	}

	// step 1: copy everything from source to dst if src newer
	// for file in filesetSrc: src file exists in dst ?
	err = filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == src { // skip basepath
				return nil
			}
			if c.IgnoreHidden && strings.Contains(path, "/.") {
				c.Log.Debug().Msgf("skip   : hidden file '%s'", path)
				return nil
			}
			childPath := strings.TrimPrefix(path, filesetSrc.Basepath)

			// populate filesetSrc on the fly...
			filesetSrc.Paths[childPath] = srcInfo

			dstPath := filepath.Join(dst, childPath)
			if !srcInfo.IsDir() && !srcInfo.Mode().IsRegular() {
				c.Log.Debug().Msgf("skip non-regular file '%v'", path)
				return nil
			}

			if filesetDst.Contains(childPath) {
				//   yes --> compare, copy if newer
				c.Log.Debug().Msgf("file '%v' exists in dst, compare", childPath)
				dstInfo, _ := os.Stat(filepath.Join(filesetDst.Basepath, childPath))
				if compare.BasicUnequal(srcInfo, dstInfo) {
					c.Log.Debug().Msg("  --> file in src newer, copy")
					err := copyFileOrCreateDir(
						filepath.Join(filesetSrc.Basepath, childPath),
						filepath.Join(filesetDst.Basepath, childPath),
						srcInfo,
						c,
					)
					if err != nil {
						c.Log.Error().Err(err).Msg("copy failed")
					}
					return nil
				} else {
					c.Log.Debug().Msg("  --> file in src not newer, skip")
					return nil
				}
			} else {
				//   no --> copy
				c.Log.Debug().Msgf("file '%v' does not exist in dst, copy", childPath)
				err := copyFileOrCreateDir(
					filepath.Join(filesetSrc.Basepath, childPath),
					filepath.Join(filesetDst.Basepath, childPath),
					srcInfo,
					c,
				)
				if err != nil {
					c.Log.Error().Err(err).Msg("copy failed")
				}
			}

			c.Log.Debug().Msgf("copy / create dir '%v'", dstPath)
			return copyFileOrCreateDir(path, dstPath, srcInfo, c)
		},
	)
	if err != nil {
		return err
	}

	// step 2: clean everything from dst that is not in src
	if !c.CleanDst {
		c.Log.Debug().Msg("config.CleanDst is false, mirror done")
		return nil
	}
	// for file in filesetDst: file exists in filesetSrc ? --> Delete if not.
	for name, dstInfo := range filesetDst.Paths {
		if !filesetSrc.Contains(name) {
			c.Log.Debug().Msgf("file '%v' does not exist in src, delete", name)
			err := deleteFileOrDir(filepath.Join(filesetDst.Basepath, name), dstInfo, c)
			if err != nil {
				c.Log.Error().Err(err).Msg("deletion failed")
				// this can sometimes give an error if the parent directory was deleted before...
			}
		}
	}
	return nil
}

// Sync synchronizes files between src and dst, keeping only the newer versions
func Sync(src, dst string, c *config.Config) error {
	c.Log.Debug().Msg("SYNC")
	c.CleanDst = false
	// step 1: copy everything from source to dst if src newer or file does not exist dst
	err := Mirror(src, dst, c)
	if err != nil {
		return err
	}
	// step 2: copy everything from dst to src if dst newer or file does not exist in src
	err = Mirror(dst, src, c)
	return err
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

func deleteFileOrDir(dst string, dstInfo fs.FileInfo, c *config.Config) error {
	if c.DryRun {
		return nil
	}
	if dstInfo.IsDir() {
		return os.RemoveAll(dst)
	}
	return os.Remove(dst)
}
