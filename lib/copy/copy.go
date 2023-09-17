package copy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
)

const (
	BUFFERSIZE                  = 4096
	DefaultModeDir  fs.FileMode = 0755 // rwxr-xr-x
	DefaultModeFile fs.FileMode = 0644 // rw-r--r--
)

// createDir wraps os.MkdirAll and ignores dir exists error
func CreateDir(dst string, dry bool) error {
	if dry {
		return nil
	}
	if err := os.MkdirAll(dst, DefaultModeDir); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// CopyFile copies src to dst. If dst exists, it will be overwritten.
// mtime and atime of the destination file will be set to that of the source file
func CopyFile(src, dst string, sourceFileStat fs.FileInfo, dry bool) error {
	if dry {
		return nil
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("'%s' is not a regular file", src)
	}

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

	if err := os.Chtimes(dst, sourceFileStat.ModTime(), sourceFileStat.ModTime()); err != nil {
		return nil
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

func ByteCount(b uint) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func DeleteFileOrDir(dst string, dstInfo fs.FileInfo, dry bool) error {
	if dry {
		return nil
	}
	if dstInfo.IsDir() {
		return os.RemoveAll(dst)
	}
	return os.Remove(dst)
}
