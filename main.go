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

	"gosyncit/lib/compare"
	cp "gosyncit/lib/copy"
	"gosyncit/lib/pathlib"
)

const (
	defaultModeDir  fs.FileMode = 0755 // rwxr-xr-x
	defaultModeFile fs.FileMode = 0644 // rw-r--r--
)

type config struct {
	// config string // path to a config file (toml) // TODO : export config fields
	dryRun          bool // only print to console
	ignoreHidden    bool // ignore ".*"
	forceWrite      bool // overwrite files in dst even if newer
	cleanDst        bool // remove files from dst that do not exist in src
	keepPermissions bool // set original permissions for updated files
	deepCompare     bool // compare files on a byte level if basic comparison (mtime, size) returns true
}

func (c *config) String() string {
	repr := fmt.Sprintf("dryRun: %v", c.dryRun)
	repr += fmt.Sprintf(", ignoreHidden: %v", c.ignoreHidden)
	repr += fmt.Sprintf(", forceWrite: %v", c.forceWrite)
	repr += fmt.Sprintf(", cleanDst: %v", c.cleanDst)
	repr += fmt.Sprintf(", keepPermissions: %v", c.keepPermissions)
	repr += fmt.Sprintf(", deepCompare: %v", c.deepCompare)
	return repr
}

func (c *config) load(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	var (
		dryRun       = flags.Bool("dryRun", true, "dry run: only tell what you will do but don't actually do it")
		ignoreHidden = flags.Bool("ignoreHidden", true, "ignore hidden files")
		forceWrite   = flags.Bool("forceWrite", false, "force write from source to destination, regardless of file timestamp etc.")
		cleanDst     = flags.Bool("cleanDst", false, "strict mirror, only files from source must exist in destination")
		deepCompare  = flags.Bool("deepCompare", false, "compare files on a byte level if basic comparison (mtime, size) returns true")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c.dryRun = *dryRun
	c.ignoreHidden = *ignoreHidden // TODO : add include / exclude filters, --> config file ?!
	c.forceWrite = *forceWrite
	c.cleanDst = *cleanDst
	c.keepPermissions = true // TODO : this does not work on windows ?!
	c.deepCompare = *deepCompare

	return nil
}

func handleErrFatal(err error) {
	if err != nil {
		slog.Error("fatal", err)
		os.Exit(1)
	}
}

// mirror src to dst. mirroring is done "loose", i.e. files that exist in the destination
// but not in the source are kept.
func mirror(src, dst string, c *config) error {
	if src == dst {
		return errors.New("source and destination path are identical")
	}

	slog.Info(fmt.Sprintf("config : %s\n---", c))

	err := filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if c.ignoreHidden {
				if strings.Contains(path, "/.") {
					slog.Info(fmt.Sprintf("skip   : hidden file '%s'", path))
					return nil
				}
			}
			childPath := strings.TrimPrefix(path, src) // for the root dir, childPath will be empty
			dstPath := filepath.Join(dst, childPath)
			dstInfo, err := os.Stat(dstPath)

			// destination does not exist, create directory or copy file
			if os.IsNotExist(err) {
				if srcInfo.IsDir() {
					slog.Info(fmt.Sprintf("create : create '%s'", dstPath))
					return copyFileOrCreateDir(path, dstPath, srcInfo, c)
				}
				if !srcInfo.Mode().IsRegular() {
					slog.Info(fmt.Sprintf("skip   : '%s' is not a regular file", path))
					return nil
				}
				slog.Info(fmt.Sprintf("create : file '%s'", dstPath))
				return copyFileOrCreateDir(path, dstPath, srcInfo, c)
			}

			// destination exists; compare files, skip directories
			if err == nil {
				if srcInfo.IsDir() {
					slog.Info(fmt.Sprintf("skip   : exists and is a directory: '%s'", dstPath))
					return nil
				}
				// TODO : add deep-compare option
				if compare.BasicEqual(srcInfo, dstInfo) || c.forceWrite {
					slog.Info("copy   : source file changed or overwrite enforced")
					return copyFileOrCreateDir(path, dstPath, srcInfo, c)
				}
				slog.Info(fmt.Sprintf("skip   : source file '%s' older or equal", path))
				return nil
			}
			return err
		},
	)

	return err
}

func copyFileOrCreateDir(src, dst string, sourceFileStat fs.FileInfo, c *config) error {
	if c.dryRun {
		return nil
	}
	if sourceFileStat.IsDir() {
		return os.MkdirAll(dst, defaultModeDir)
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	err := cp.CopyFile(src, dst)
	if err != nil {
		return err
	}
	if c.keepPermissions && runtime.GOOS != "windows" {
		return os.Chmod(dst, sourceFileStat.Mode())
	}
	return nil
}

func main() {
	c := &config{}
	err := c.load(os.Args)
	handleErrFatal(err)

	wd, _ := os.Getwd()
	src := filepath.Join(wd, "testdata/dirA/")
	src, err = pathlib.CheckPath(src)
	handleErrFatal(err)

	dst := filepath.Join(wd, "testdata/dirB/")
	dst, err = pathlib.CheckPath(dst)
	handleErrFatal(err)

	err = mirror(src, dst, c)
	handleErrFatal(err)
}
