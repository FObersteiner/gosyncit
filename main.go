package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultModeDir  fs.FileMode = 0755 // rwxr-xr-x
	defaultModeFile fs.FileMode = 0644 // rw-r--r--
	BUFFERSIZE                  = 4096 // is also default ext4 sector size
)

type config struct {
	dryRun          bool // only print to console
	ignoreHidden    bool // ignore ".*"
	forceWrite      bool // overwrite files in dst even if newer
	cleanDst        bool // remove files from dst that do not exist in src
	keepPermissions bool // set original premissions for updated files
}

func (c *config) load() error {
	// TODO : populate from flags
	c.dryRun = true
	c.ignoreHidden = true
	// TODO : add include / exclude filters, --> config file ?!
	c.forceWrite = false
	c.cleanDst = false

	// TODO : this does not work on windows:
	c.keepPermissions = true
	// --> notify the user

	return nil
}

func handleErrFatal(err error, log *zerolog.Logger) {
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
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

// mirror src to dst
func mirror(src, dst string, c *config, log *zerolog.Logger) error {
	if src == dst {
		return errors.New("source and destination are identical")
	}
	// pt.1 : src --> dst
	err := filepath.Walk(src,
		func(path string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if c.ignoreHidden {
				if strings.Contains(path, "/.") {
					return nil
				}
			}
			childPath := strings.TrimPrefix(path, src) // for the root dir, childPath will be empty
			dstPath := filepath.Join(dst, childPath)

			dstInfo, err := os.Stat(dstPath)

			// destination does not exist, create directory or copy file
			if os.IsNotExist(err) {
				log.Info().Msgf("create : directory or file %v", dstPath)
				if c.dryRun {
					return nil
				}
				return copyFileOrCreateDir(path, dstPath, srcInfo, c, log)
			}

			// destination exists; compare files, skip directories
			if err == nil {
				log.Info().Msgf("check  : directory or file %v", dstPath)
				if srcInfo.IsDir() {
					log.Info().Msg("skip   : destination exists and is a directory, skipping")
					return nil
				}
				if srcInfo.ModTime().After(dstInfo.ModTime()) || c.forceWrite {
					log.Info().Msg("copy   : source file newer or overwrite enforced")
					if c.dryRun {
						return nil
					}
					return copyFileOrCreateDir(path, dstPath, srcInfo, c, log)
				}
				log.Info().Msg("skip   : source file not newer")
				return nil
			}
			return err
		},
	)
	if err != nil {
		return err
	}
	// pt.2 : if dst should be an exact mirror, remove everything from dst that is not in src

	return nil
}

func copyFileOrCreateDir(src, dst string, sourceFileStat fs.FileInfo, c *config, log *zerolog.Logger) error {
	if sourceFileStat.IsDir() {
		return os.MkdirAll(dst, defaultModeDir)
	}
	if !sourceFileStat.Mode().IsRegular() {
		log.Info().Msgf("skip   : %s is not a regular file", src)
		// TODO : follow symlinks ? --> os.Readlink
		return nil
	}
	err := copyFile(src, dst)
	if err != nil {
		return err
	}
	if c.keepPermissions && runtime.GOOS != "windows" {
		srcMode := sourceFileStat.Mode()
		return os.Chmod(dst, srcMode)
	}
	return nil
}

func copyFile(src, dst string) error {
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

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	c := &config{}
	err := c.load()
	handleErrFatal(err, &log)

	src := "/home/floo/Code/go/dirfun/testdata/dirA/"
	src, err = checkPath(src)
	handleErrFatal(err, &log)

	dst := "/home/floo/Code/go/dirfun/testdata/dirB/"
	dst, err = checkPath(dst)
	handleErrFatal(err, &log)

	err = mirror(src, dst, c, &log)
	handleErrFatal(err, &log)
}
