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
)

// sftpsyncCmd represents the sftpsync command
var sftpsyncCmd = &cobra.Command{
	Use:     "sftpsync 'server-url' 'src' 'dst'",
	Aliases: []string{"ssy"},
	Short:   "sync directories via SFTP",
	// Long: "",
	Args: cobra.MaximumNArgs(3),
	RunE: func(_ *cobra.Command, args []string) error {
		url := viper.GetString("url")
		src := viper.GetString("src")
		dst := viper.GetString("dst")
		if len(args) < 3 && (url == "" || src == "" || dst == "") {
			return errors.New("missing required argument 'URL', 'src' or 'dst'")
		}

		if len(args) == 3 {
			url = args[0]
			src = args[1]
			dst = args[2]
		}

		dry := viper.GetBool("dryrun")
		ignorehidden := viper.GetBool("skiphidden")
		setGlobalVerbose := viper.GetBool("verbose")
		verbose = setGlobalVerbose
		port := 22
		user := "me"

		return SftpSync(url, user, src, dst, port, dry, ignorehidden)
	},
}

func init() {
	rootCmd.AddCommand(sftpsyncCmd)
	sftpsyncCmd.Flags().SortFlags = false

	sftpsyncCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", false, "show what will be done")
	err := viper.BindPFlag("dryrun", sftpsyncCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to 'dryrun' flag:", err)
	}

	sftpsyncCmd.Flags().BoolVarP(&skipHidden, "skiphidden", "s", false, "skip hidden files")
	err = viper.BindPFlag("skiphidden", sftpsyncCmd.Flags().Lookup("skiphidden"))
	if err != nil {
		log.Fatal("error binding viper to 'skiphidden' flag:", err)
	}

	sftpsyncCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output to the command line")
	err = viper.BindPFlag("verbose", sftpsyncCmd.Flags().Lookup("verbose"))
	if err != nil {
		log.Fatal("error binding viper to 'verbose' flag:", err)
	}
}

// SftpSync mirrors directory 'src' to 'dst' via SFTP
func SftpSync(url, usr, src, dst string, port int, dry, ignorehidden bool) error {
	log.Println(url, port, usr)
	log.Println(src, dst)
	log.Println(dry, ignorehidden)
	return nil
}
