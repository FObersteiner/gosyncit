package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"log/slog"

	cp "gosyncit/lib/copy"
)

const (
	defaultModeDir  fs.FileMode = 0755 // rwxr-xr-x
	defaultModeFile fs.FileMode = 0644 // rw-r--r--
)

type config struct {
	dryRun          bool // only print to console
	ignoreHidden    bool // ignore ".*"
	forceWrite      bool // overwrite files in dst even if newer
	cleanDst        bool // remove files from dst that do not exist in src
	keepPermissions bool // set original permissions for updated files
}

func (c *config) String() string {
	repr := fmt.Sprintf("dryRun: %v", c.dryRun)
	repr += fmt.Sprintf(", ignoreHidden: %v", c.ignoreHidden)
	repr += fmt.Sprintf(", forceWrite: %v", c.forceWrite)
	repr += fmt.Sprintf(", cleanDst: %v", c.cleanDst)
	repr += fmt.Sprintf(", keepPermissions: %v", c.keepPermissions)
	return repr
}

func (c *config) load(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	var (
		dryRun       = flags.Bool("dryRun", true, "dry run: only tell what you will do but don't actually do it")
		ignoreHidden = flags.Bool("ignoreHidden", true, "ignore hidden files")
		forceWrite   = flags.Bool("forceWrite", false, "force write from source to destination, regardless of file timestamp etc.")
		cleanDst     = flags.Bool("cleanDst", false, "strict mirror, only files from source must exist in destination")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c.dryRun = *dryRun
	c.ignoreHidden = *ignoreHidden // TODO : add include / exclude filters, --> config file ?!
	c.forceWrite = *forceWrite
	c.cleanDst = *cleanDst
	c.keepPermissions = true // TODO : this does not work on windows ?!

	return nil
}

func handleErrFatal(err error) {
	if err != nil {
		slog.Error("fatal", err)
		os.Exit(1)
	}
}

func checkPath(path string) (string, error) {
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

// mirror src to dst. mirroring is done "loose", i.e. files that exist in the destination
// but not in the source are kept.
func mirror(src, dst string, c *config) error {
	if src == dst {
		return errors.New("source and destination are identical")
	}

	slog.Info(fmt.Sprintf("config : %s", c))

	err := filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if c.ignoreHidden {
				if strings.Contains(path, "/.") {
					slog.Info(fmt.Sprintf("skip   : hidden file %v", path))
					return nil
				}
			}
			childPath := strings.TrimPrefix(path, src) // for the root dir, childPath will be empty
			dstPath := filepath.Join(dst, childPath)
			dstInfo, err := os.Stat(dstPath)

			// destination does not exist, create directory or copy file
			if os.IsNotExist(err) {
				if !srcInfo.Mode().IsRegular() {
					slog.Info(fmt.Sprintf("skip   : %s is not a regular file", src))
					return nil
				}
				slog.Info(fmt.Sprintf("create : directory or file %v", dstPath))
				if c.dryRun {
					return nil
				}
				return copyFileOrCreateDir(path, dstPath, srcInfo, c)
			}

			// destination exists; compare files, skip directories
			if err == nil {
				slog.Info(fmt.Sprintf("check  : directory or file %v", dstPath))
				if srcInfo.IsDir() {
					slog.Info("skip   : destination exists and is a directory")
					return nil
				}
				if srcInfo.ModTime().After(dstInfo.ModTime()) || c.forceWrite {
					slog.Info("copy   : source file newer or overwrite enforced")
					if c.dryRun {
						return nil
					}
					return copyFileOrCreateDir(path, dstPath, srcInfo, c)
				}
				slog.Info("skip   : source file not newer")
				return nil
			}
			return err
		},
	)

	return err
}

func copyFileOrCreateDir(src, dst string, sourceFileStat fs.FileInfo, c *config) error {
	if sourceFileStat.IsDir() {
		return os.MkdirAll(dst, defaultModeDir)
	}
	if !sourceFileStat.Mode().IsRegular() {
		slog.Info(fmt.Sprintf("skip   : %s is not a regular file", src))
		return nil
	}
	err := cp.CopyFile(src, dst)
	if err != nil {
		return err
	}
	if c.keepPermissions && runtime.GOOS != "windows" {
		srcMode := sourceFileStat.Mode()
		return os.Chmod(dst, srcMode)
	}
	return nil
}

func main() {
	c := &config{}
	err := c.load(os.Args)
	handleErrFatal(err)

	wd, _ := os.Getwd()
	src := filepath.Join(wd, "testdata/dirA/")
	src, err = checkPath(src)
	handleErrFatal(err)

	dst := filepath.Join(wd, "testdata/dirB/")
	dst, err = checkPath(dst)
	handleErrFatal(err)

	err = mirror(src, dst, c)
	handleErrFatal(err)
}
