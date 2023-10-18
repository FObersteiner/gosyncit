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

var syncCmd = &cobra.Command{
	Use:     "sync 'src' 'dst'",
	Aliases: []string{"sy"},
	Short:   "synchronize directory 'src' with directory 'dst'",
	Long: `Synchronize the content of source directory with destination directory.
Files will be copied if one is newer or doesn't exit in the destination.`,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		src := viper.GetString("src")
		dst := viper.GetString("dst")
		if len(args) < 2 && (src == "" || dst == "") {
			return errors.New("missing required argument 'src' or 'dst'")
		}

		if len(args) == 2 {
			src = args[0]
			dst = args[1]
		}

		dry := viper.GetBool("dryrun")
		ignorehidden := viper.GetBool("skiphidden")
		setGlobalVerbose := viper.GetBool("verbose")
		verbose = setGlobalVerbose

		return Sync(src, dst, dry, ignorehidden)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().SortFlags = false

	syncCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", false, "show what will be done")
	err := viper.BindPFlag("dryrun", syncCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to 'dryrun' flag:", err)
	}

	syncCmd.Flags().BoolVarP(&skipHidden, "skiphidden", "s", false, "skip hidden files")
	err = viper.BindPFlag("skiphidden", syncCmd.Flags().Lookup("skiphidden"))
	if err != nil {
		log.Fatal("error binding viper to 'skiphidden' flag:", err)
	}

	syncCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output to the command line")
	err = viper.BindPFlag("verbose", syncCmd.Flags().Lookup("verbose"))
	if err != nil {
		log.Fatal("error binding viper to 'verbose' flag:", err)
	}
}

// ------------------------------------------------------------------------------------

// Sync synchronizes directory 'src' with directory 'dst'.
func Sync(src, dst string, dry bool, skipHidden bool) error {
	fmt.Println("~~~ SYNC ~~~")
	fmt.Printf("'%s' <--> '%s'\n\n", src, dst)

	var nItems, nBytes uint
	t0 := time.Now()

	src, dst, err := pathlib.CheckSrcDst(src, dst)
	if err != nil {
		verboseprint("path check error:", err)
		return err
	}

	filesetSrc, err := fileset.New(src)
	if err != nil {
		verboseprint("src file set creation error:", err)
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
		verboseprint("dst fileset population got error", err)
		if !dry {
			verboseprint("dst might not exist, try to create.")
			err := os.MkdirAll(dst, copy.DefaultModeDir)
			if err != nil {
				return err
			}
		}
	}

	// we also need a 'seen' map to track which files were copied from src to dst,
	// so we can skip copying them from dst to src (as their mtime will be newer)
	newInDst := make(map[string]struct{})

	basepath := strings.TrimSuffix(filesetSrc.Basepath, string(os.PathSeparator))

	// STEP 1 : copy everything from source to dst if src newer
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
				verboseprintf("skip hidden '%s'\n", srcPath)
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
				verboseprintf("create or skip dir '%s'\n", dstPath)
				return copy.CreateDir(dstPath, dry) // ignores error if dir exists
			}

			if !srcInfo.Mode().IsRegular() {
				verboseprintf("skip non-regular file '%s'\n", srcPath)
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
			if compare.SrcYounger(srcInfo, dstInfo) {
				fmt.Printf("overwrite file (src -> dst) '%s'\n", srcPath)
				newInDst[childPath] = struct{}{}
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			} else {
				verboseprintf("skip file '%s'\n", srcPath)
			}
			return nil
		},
	)

	if err != nil {
		verboseprint("src filepath walk error:", err)
		return err
	}

	// STEP 2 : copy everything from dst to src if dst newer
	err = filepath.Walk(dst,
		func(srcPath string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			childPath := strings.TrimPrefix(srcPath, filesetDst.Basepath)
			if childPath == basepath {
				return nil // skip basepath
			}

			if skipHidden && (strings.HasPrefix(srcPath, ".") || strings.Contains(srcPath, "/.")) {
				verboseprintf("skip hidden '%s'\n", srcPath)
				return nil
			}

			nItems++
			nBytes += uint(srcInfo.Size())

			dstPath := filepath.Join(src, childPath)

			// A) item is directory.
			//   exists in dst?
			//     no  --> create.
			//     yes --> skip.
			if srcInfo.IsDir() {
				verboseprintf("create or skip dir '%s'\n", dstPath)
				return copy.CreateDir(dstPath, dry) // ignores error if dir exists
			}

			if !srcInfo.Mode().IsRegular() {
				verboseprintf("skip non-regular file '%s'\n", srcPath)
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
				verboseprintf("skip new file '%s'\n", srcPath)
				return nil
			}

			dstInfo, _ := os.Stat(filepath.Join(filesetSrc.Basepath, childPath))
			if compare.SrcYounger(srcInfo, dstInfo) {
				fmt.Printf("overwrite file (dst -> src) '%s'\n", srcPath)
				return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
			} else {
				verboseprintf("skip file '%s'\n", srcPath)
			}
			return nil
		},
	)

	if err != nil {
		verboseprint("dst filepath walk error:", err)
		return err
	}

	dt := time.Since(t0)
	fmt.Printf("\n~~~ SYNC done ~~~\n%v items, %v, in %v\n~~~\n",
		nItems,
		copy.ByteCount(nBytes),
		dt,
	)
	return nil
}
