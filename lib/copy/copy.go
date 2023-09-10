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
func SimpleCopy(src, dst string, dry, clean bool) error {
	fmt.Println("Called SimpelCopy, dry, clean:", dry, clean)
	src, dst, err := pathlib.CheckSrcDst(src, dst)
	if err != nil {
		fmt.Println("checkpath error")
		return err
	}

	if clean {
		fmt.Println("deleting dst for a clean copy...")
		if !dry {
			err := os.RemoveAll(dst)
			if err != nil {
				return err
			}
		}
	}

	return filepath.Walk(src,
		func(srcPath string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			childPath := strings.TrimPrefix(srcPath, src)
			dstPath := filepath.Join(dst, childPath)

			if srcInfo.IsDir() {
				fmt.Printf("create dir '%s'\n", dstPath)
				return createDir(dstPath, dry)
			}

			if !srcInfo.Mode().IsRegular() {
				fmt.Printf("skip non-regular file '%s'\n", srcPath)
				return nil
			}

			// SimpleCopy ignores existing files:
			fmt.Printf("copy file '%s'\n", srcPath)
			return CopyFile(srcPath, dstPath, srcInfo, dry)

			// TODO : mirror
			// A) directory.
			//   exists in dst?
			//     no  --> create.
			//     yes --> skip.
			// B)
			//   exists in dst?
			//     no  --> write.
			//     yes --> overwrite?
			//       yes --> write.
			//       no  --> src younger?
			//         yes --> write.
			//         no  --> skip.
		},
	)
}

// createDir wraps os.MkdirAll and ignores dir exists error
func createDir(dst string, dry bool) error {
	if dry {
		return nil
	}
	if err := os.MkdirAll(dst, defaultModeDir); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// CopyFile copies src to dst. If dst exists, it will be overwritten.
func CopyFile(src, dst string, sourceFileStat fs.FileInfo, dry bool) error {
	if dry {
		return nil
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
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
