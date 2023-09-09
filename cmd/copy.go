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
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gosyncit/lib/copy"
)

var dryRun bool

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:     "copy",
	Aliases: []string{"cp"},
	Short:   "copy src to dst",
	Long: `Copy the content of source directory to destination directory.
  If destination exists, any existing content will be ignored.`,
	Args:      cobra.MaximumNArgs(2),
	ValidArgs: []string{"src", "dst"},
	RunE: func(cmd *cobra.Command, args []string) error {
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
		return copy.SimpleCopy(src, dst, dry)
	},
}

func init() {
	copyCmd.Flags().SortFlags = false

	copyCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", true, "show what will be done") // same as rsync
	err := viper.BindPFlag("dryrun", copyCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to dryrun flag:", err)
	}

	rootCmd.AddCommand(copyCmd)
}
