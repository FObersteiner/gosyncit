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
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FObersteiner/gosyncit/lib/fileset"
	libsftp "github.com/FObersteiner/gosyncit/lib/sftp"
)

// sftpmirrorCmd represents the sftpsync command
var sftpmirrorCmd = &cobra.Command{
	Use:     "sftpmirror 'local' 'remote' 'remote-url' 'username'",
	Aliases: []string{"smir"},
	Short:   "mirrors directories via SFTP",
	Long: `the direction can either be local --> remote or remote --> local.
  "local" in this context means local file system, remote means file system of the sftp server.`,
	Args: cobra.MaximumNArgs(4),
	RunE: func(_ *cobra.Command, args []string) error {
		local := viper.GetString("local")
		remote := viper.GetString("remote")
		url := viper.GetString("remote-url")
		usr := viper.GetString("username")

		if len(args) < 4 && (url == "" || local == "" || remote == "") {
			return errors.New("missing required argument 'local', 'remote', 'URL' or 'username'")
		}

		if len(args) == 4 {
			local = args[0]
			remote = args[1]
			url = args[2]
			usr = args[3]
		}

		p := viper.GetInt("port")

		dry := viper.GetBool("dryrun")
		ignorehidden := viper.GetBool("skiphidden")
		setGlobalVerbose := viper.GetBool("verbose")
		verbose = setGlobalVerbose

		creds := libsftp.Credentials{
			Usr:        usr,
			Host:       url,
			AgentSock:  "SSH_AUTH_SOCK",
			Port:       p,
			SSHtimeout: time.Second * 10,
		}

		return SftpMir(local, remote, creds, dry, ignorehidden)
	},
}

func init() {
	rootCmd.AddCommand(sftpmirrorCmd)
	sftpmirrorCmd.Flags().SortFlags = false

	sftpmirrorCmd.Flags().IntVarP(&port, "port", "p", 22, "ssh port number")
	err := viper.BindPFlag("port", sftpmirrorCmd.Flags().Lookup("port"))
	if err != nil {
		log.Fatal("error binding viper to 'port' flag:", err)
	}

	sftpmirrorCmd.Flags().BoolVarP(&dryRun, "dryrun", "n", false, "show what will be done")
	err = viper.BindPFlag("dryrun", sftpmirrorCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal("error binding viper to 'dryrun' flag:", err)
	}

	sftpmirrorCmd.Flags().BoolVarP(&skipHidden, "skiphidden", "s", false, "skip hidden files")
	err = viper.BindPFlag("skiphidden", sftpmirrorCmd.Flags().Lookup("skiphidden"))
	if err != nil {
		log.Fatal("error binding viper to 'skiphidden' flag:", err)
	}

	sftpmirrorCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output to the command line")
	err = viper.BindPFlag("verbose", sftpmirrorCmd.Flags().Lookup("verbose"))
	if err != nil {
		log.Fatal("error binding viper to 'verbose' flag:", err)
	}
}

// SftpMir mirrors directory 'local' to 'remote' via SFTP
func SftpMir(local, remote string, creds libsftp.Credentials, dry, ignorehidden bool) error {
	fmt.Println("~~~ SFTP MIRROR ~~~")
	verboseprintf("%s\n", &creds)
	fmt.Printf("'%s' --> '%s'\n\n", local, remote)

	sshcon, err := libsftp.GetSSHconn(creds)
	if err != nil {
		return err
	}
	defer sshcon.Close()

	sc, err := sftp.NewClient(sshcon)
	if err != nil {
		return err
	}
	defer sc.Close()
	verboseprintf("SFTP connection established; %s", &creds)

	// TODO : handle non-existing cases (local, remote)

	// TODO : handle reverse-case: mirror remote to local

	filesetLocal, err := fileset.New(local)
	if err != nil {
		verboseprint("local file set creation error:", err)
		return err
	}
	err = filesetLocal.Populate()
	if err != nil {
		verboseprint("local fileset population got error", err)
		return err
	}

	if !strings.HasSuffix(remote, string(os.PathSeparator)) {
		remote += string(os.PathSeparator)
	}
	filesetRemote := fileset.Fileset{
		Basepath: remote,
		Paths:    make(map[string]fs.FileInfo),
	}

	err = filesetRemote.SftpPopulate(sc)
	if err != nil {
		verboseprint("remote fileset population got error", err)
		return err
	}

	fmt.Println(filesetLocal.String())
	fmt.Println(filesetRemote.String())

	return nil
}
