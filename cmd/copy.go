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
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FObersteiner/gosyncit/lib/copy"
	"github.com/FObersteiner/gosyncit/lib/pathlib"
)

var copyCmd = &cobra.Command{
	Use:     "copy",
	Aliases: []string{"cp"},
	Short:   "copy directory 'src' to directory 'dst'",
	Long: `Copy the content of source directory to destination directory.
  If destination exists, any existing content will be ignored (overwritten).`,
	Args:      cobra.MaximumNArgs(2),
	ValidArgs: []string{"src", "dst"},
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
		clean := !viper.GetBool("dirty")

		return SimpleCopy(src, dst, dry, clean)
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	copyCmd.Flags().SortFlags = false

	copyCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", false, "show what will be done") // same as rsync
	err := viper.BindPFlag("dryrun", copyCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to 'dryrun' flag:", err)
	}

	copyCmd.Flags().BoolVarP(&noCleanDst, "dirty", "x", false, "do not remove anything from dst before copy")
	err = viper.BindPFlag("dirty", copyCmd.Flags().Lookup("dirty"))
	if err != nil {
		log.Fatal("error binding viper to 'dirty' flag:", err)
	}
}

// ------------------------------------------------------------------------------------

// SimpleCopy makes a copy of directory 'src' to directory 'dst'.
func SimpleCopy(src, dst string, dry, clean bool) error {
	fmt.Println("~~~ COPY ~~~")

	var nItems, nBytes uint
	t0 := time.Now()

	src, dst, err := pathlib.CheckSrcDst(src, dst)
	if err != nil {
		fmt.Println("path check: error")
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

	err = filepath.Walk(src,
		func(srcPath string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			nItems++
			nBytes += uint(srcInfo.Size())

			childPath := strings.TrimPrefix(srcPath, src)
			dstPath := filepath.Join(dst, childPath)

			if srcInfo.IsDir() {
				fmt.Printf("create dir '%s'\n", dstPath)
				return copy.CreateDir(dstPath, dry)
			}

			if !srcInfo.Mode().IsRegular() {
				fmt.Printf("skip non-regular file '%s'\n", srcPath)
				return nil
			}

			// SimpleCopy ignores existing files:
			fmt.Printf("copy file '%s'\n", srcPath)
			return copy.CopyFile(srcPath, dstPath, srcInfo, dry)
		},
	)

	if err != nil {
		fmt.Println("src filepath walk error:", err)
		return err
	}

	dt := time.Since(t0)
	secs := float64(dt) / float64(time.Second)
	fmt.Printf("~~~ COPY done ~~~\n%v items (%v) in %v, %v per second\n~~~\n",
		nItems,
		copy.ByteCount(nBytes),
		dt,
		copy.ByteCount(uint(float64(nBytes)/secs)),
	)
	return nil
}
