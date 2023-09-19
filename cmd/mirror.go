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

	"github.com/FObersteiner/gosyncit/lib/compare"
	"github.com/FObersteiner/gosyncit/lib/copy"
	"github.com/FObersteiner/gosyncit/lib/fileset"
	"github.com/FObersteiner/gosyncit/lib/pathlib"
)

var mirrorCmd = &cobra.Command{
	Use:     "mirror",
	Aliases: []string{"mi"},
	Short:   "mirror directory 'src' to directory 'dst'",
	Long: `Mirror the content of source directory to destination directory.
Files will only be copied if the source file is newer.
By default, anything that exists in the destination but not in the source will be deleted.`,
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
		clean := viper.GetBool("clean")
		ignorehidden := viper.GetBool("skiphidden")

		return Mirror(src, dst, dry, clean, ignorehidden)
	},
}

func init() {
	rootCmd.AddCommand(mirrorCmd)

	mirrorCmd.Flags().SortFlags = false

	mirrorCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", false, "show what will be done")
	err := viper.BindPFlag("dryrun", mirrorCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to 'dryrun' flag:", err)
	}

	mirrorCmd.Flags().BoolVarP(&cleanDst, "clean", "x", true, "remove everything from dst that is not found in source")
	err = viper.BindPFlag("clean", mirrorCmd.Flags().Lookup("clean"))
	if err != nil {
		log.Fatal("error binding viper to 'clean' flag:", err)
	}

	mirrorCmd.Flags().BoolVarP(&skipHidden, "skiphidden", "s", false, "skip hidden files")
	err = viper.BindPFlag("skiphidden", mirrorCmd.Flags().Lookup("skiphidden"))
	if err != nil {
		log.Fatal("error binding viper to 'skiphidden' flag:", err)
	}
}

// ------------------------------------------------------------------------------------

// Mirror mirrors directory 'src' to directory 'dst'.
func Mirror(src, dst string, dry, clean, skipHidden bool) error {

	fmt.Println("~~~ MIRROR ~~~")
	fmt.Printf("'%s' <--> '%s'\n\n", src, dst)

	var nItems, nBytes uint
	t0 := time.Now()

	src, dst, err := pathlib.CheckSrcDst(src, dst)
	if err != nil {
		fmt.Println("path check error:", err)
		return err
	}

	filesetSrc, err := fileset.New(src)
	if err != nil {
		fmt.Println("src file set creation error:", err)
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
			err := os.MkdirAll(dst, copy.DefaultModeDir)
			if err != nil {
				return err
			}
		}
	}

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

			if skipHidden && (strings.HasPrefix(srcPath, ".") || strings.Contains(srcPath, "/.")) {
				fmt.Printf("skip hidden '%s'\n", srcPath)
				return nil
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
				fmt.Printf("overwrite file '%s'\n", srcPath)
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			} else {
				fmt.Printf("skip file '%s'\n", srcPath)
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	// fmt.Println("\nfileSet SRC:")
	// fmt.Println(filesetSrc)

	// step 2: clean everything from dst that is not in src
	if clean {
		// for file in filesetDst: file exists in filesetSrc ? --> Delete if not.
		for name, dstInfo := range filesetDst.Paths {

			if skipHidden && (strings.HasPrefix(name, ".") || strings.Contains(name, "/.")) {
				fmt.Printf("skip hidden '%s'\n", name)
				return nil
			}

			if !filesetSrc.Contains(name) {
				fmt.Printf("file '%v' does not exist in src, delete\n", name)
				err := copy.DeleteFileOrDir(filepath.Join(filesetDst.Basepath, name), dstInfo, dry)
				if err != nil {
					// this can sometimes give an error if the parent directory was deleted before...
					fmt.Println("deletion failed,", err)
				}
			}
		}
	}

	dt := time.Since(t0)
	secs := float64(dt) / float64(time.Second)
	fmt.Printf("~~~ mirror done ~~~\n%v items (%v) in %v, %v per second\n~~~\n",
		nItems,
		copy.ByteCount(nBytes),
		dt,
		copy.ByteCount(uint(float64(nBytes)/secs)),
	)
	return nil
}
