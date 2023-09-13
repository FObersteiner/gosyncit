/*
Copyright © 2023 Florian Obersteiner <f.obersteiner@posteo.de>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gosyncit/lib/compare"
	"gosyncit/lib/copy"
	"gosyncit/lib/fileset"
	"gosyncit/lib/pathlib"
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"sy"},
	Short:   "synchronize directory 'src' with directory 'dst'",
	Long: `Synchronize the content of source directory with destination directory.
Files will be copied if one is newer or doesn't exit in the destination.`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		src := viper.GetString("src")
		dst := viper.GetString("dst")

		if len(args) < 2 && (src == "" || dst == "") {
			return errors.New("missing required arguments [src] and [dst]")
		}

		if len(args) == 2 {
			src = args[0]
			dst = args[1]
		}

		dry := viper.GetBool("dryrun")

		return Sync(src, dst, dry)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().SortFlags = false

	syncCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", true, "show what will be done") // same as rsync
	err := viper.BindPFlag("dryrun", syncCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to 'dryrun' flag:", err)
	}
}

func Sync(src, dst string, dry bool) error {
	fmt.Println("~~~ SYNC ~~~")

	var nItems, nBytes uint
	t0 := time.Now()

	src, dst, err := pathlib.CheckSrcDst(src, dst)
	if err != nil {
		fmt.Println("path check: error")
		return err
	}

	filesetSrc, err := fileset.New(src)
	if err != nil {
		fmt.Println("file set creation: error")
		return err
	}

	if !strings.HasSuffix(dst, string(os.PathSeparator)) {
		dst += string(os.PathSeparator)
	}
	filesetDst := fileset.Fileset{
		Basepath: dst,
		Paths:    make(map[string]fs.FileInfo),
	}

	// we need a fileset for the destination, to check against while walking the src
	// for file in filesetSrc: src file exists in dst ?
	err = filesetDst.Populate()
	if err != nil {
		fmt.Println("dst fileset population got error", err)
		if !dry {
			fmt.Println("dst might not exist, try to create.")
			_ = os.MkdirAll(dst, copy.DefaultModeDir)
		}
	}

	// we also need a 'seen' map to track which files were copied from src to dst,
	// so we can skip copying them from dst to src (as their mtime will be newer)
	newInDst := make(map[string]struct{})

	basepath := strings.TrimSuffix(filesetSrc.Basepath, string(os.PathSeparator))

	// step 1: copy everything from source to dst if src newer
	err = filepath.Walk(src,
		func(srcPath string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			childPath := strings.TrimPrefix(srcPath, filesetSrc.Basepath)
			if childPath == basepath {
				return nil // skip basepath
			}

			nItems++
			nBytes += uint(srcInfo.Size())

			dstPath := filepath.Join(dst, childPath)

			// populate filesetSrc on the fly...
			filesetSrc.Paths[childPath] = srcInfo

			// A) item is directory.
			//   exists in dst?
			//     no  --> create.
			//     yes --> skip.
			if srcInfo.IsDir() {
				fmt.Printf("create or skip dir '%s'\n", dstPath)
				return copy.CreateDir(dstPath, dry) // ignores error if dir exists
			}

			if !srcInfo.Mode().IsRegular() {
				fmt.Printf("skip non-regular file '%s'\n", srcPath)
				return nil
			}

			// B) item is file.
			//   exists in dst?
			//     no  --> write.
			//     yes --> overwrite?
			//       yes --> write.
			//       no  --> src younger?
			//         yes --> write.
			//         no  --> skip.
			if !filesetDst.Contains(childPath) {
				fmt.Printf("copy file '%s'\n", srcPath)
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			}

			dstInfo, _ := os.Stat(filepath.Join(filesetDst.Basepath, childPath))
			if compare.BasicUnequal(srcInfo, dstInfo) {
				fmt.Printf("overwrite file (src -> dst) '%s'\n", srcPath)
				newInDst[childPath] = struct{}{}
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			} else {
				fmt.Printf("skip file '%s'\n", srcPath)
			}
			return nil
		},
	)

	if err != nil {
		fmt.Println("src filepath walk error:", err)
		return err
	}

	// step 2: copy everything from dst to src if dst newer
	err = filepath.Walk(dst,
		func(srcPath string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			childPath := strings.TrimPrefix(srcPath, filesetDst.Basepath)
			if childPath == basepath {
				return nil // skip basepath
			}

			nItems++
			nBytes += uint(srcInfo.Size())

			dstPath := filepath.Join(src, childPath)

			// A) item is directory.
			//   exists in dst?
			//     no  --> create.
			//     yes --> skip.
			if srcInfo.IsDir() {
				fmt.Printf("create or skip dir '%s'\n", dstPath)
				return copy.CreateDir(dstPath, dry) // ignores error if dir exists
			}

			if !srcInfo.Mode().IsRegular() {
				fmt.Printf("skip non-regular file '%s'\n", srcPath)
				return nil
			}

			// B) item is file.
			//   exists in src?
			//     no  --> write.
			//     yes --> overwrite?
			//       yes --> write.
			//       no  --> dst younger?
			//         yes --> write.
			//         no  --> skip.
			if !filesetSrc.Contains(childPath) {
				fmt.Printf("copy file '%s'\n", srcPath)
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			}
			if _, ok := newInDst[childPath]; ok {
				fmt.Printf("skip new file '%s'\n", srcPath)
				return nil
			}

			dstInfo, _ := os.Stat(filepath.Join(filesetSrc.Basepath, childPath))
			if compare.BasicUnequal(srcInfo, dstInfo) {
				fmt.Printf("overwrite file (dst -> src) '%s'\n", srcPath)
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			} else {
				fmt.Printf("skip file '%s'\n", srcPath)
			}
			return nil
		},
	)

	if err != nil {
		fmt.Println("dst filepath walk error:", err)
		return err
	}

	dt := time.Since(t0)
	secs := float64(dt) / float64(time.Second)
	fmt.Printf("~~~ SYNC done ~~~\n%v items, %v, in %v (~%v per second)\n~~~\n",
		nItems,
		copy.ByteCount(nBytes),
		dt,
		copy.ByteCount(uint(float64(nBytes)/secs)),
	)
	return nil
}
