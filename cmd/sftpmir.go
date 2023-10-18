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

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FObersteiner/gosyncit/lib/compare"
	"github.com/FObersteiner/gosyncit/lib/copy"
	"github.com/FObersteiner/gosyncit/lib/fileset"
	"github.com/FObersteiner/gosyncit/lib/libsftp"
)

// sftpmirrorCmd represents the sftpsync command
var sftpmirrorCmd = &cobra.Command{
	Use:     "sftpmirror 'local-path' 'remote-path' 'remote-url' 'username'",
	Aliases: []string{"smir"},
	Short:   "mirrors directories via SFTP",
	Long: `the direction can either be "local --> remote" or "remote --> local".
  "local" in this context means local file system, remote means file system of the sftp server.`,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(4),
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
		reverse := viper.GetBool("reverse")
		dry := viper.GetBool("dryrun")
		ignorehidden := viper.GetBool("skiphidden")
		clean := !viper.GetBool("dirty")
		setGlobalVerbose := viper.GetBool("verbose")
		verbose = setGlobalVerbose

		creds := libsftp.Credentials{
			Usr:        usr,
			Host:       url,
			AgentSock:  "SSH_AUTH_SOCK",
			Port:       p,
			SSHtimeout: time.Second * 10,
		}

		return SftpMir(local, remote, creds, reverse, dry, ignorehidden, clean)
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

	sftpmirrorCmd.Flags().BoolVarP(&reverseDirection, "reverse", "r", false,
		"reverse mirror: remote to local instead of local to remote")
	err = viper.BindPFlag("reverse", sftpmirrorCmd.Flags().Lookup("reverse"))
	if err != nil {
		log.Fatal("error binding viper to 'reverse' flag:", err)
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

	sftpmirrorCmd.Flags().BoolVarP(&noCleanDst, "dirty", "x", false, "do not remove anything from dst that is not found in source")
	err = viper.BindPFlag("dirty", sftpmirrorCmd.Flags().Lookup("dirty"))
	if err != nil {
		log.Fatal("error binding viper to 'dirty' flag:", err)
	}

	sftpmirrorCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output to the command line")
	err = viper.BindPFlag("verbose", sftpmirrorCmd.Flags().Lookup("verbose"))
	if err != nil {
		log.Fatal("error binding viper to 'verbose' flag:", err)
	}
}

// SftpMir mirrors directory 'local' to 'remote' via SFTP
func SftpMir(local, remote string, creds libsftp.Credentials, reverse, dry, ignorehidden, clean bool) error {
	//
	// TODO : handle non-existing cases for local, remote
	// TODO : handle reverse-case: mirror remote to local
	//
	fmt.Println("~~~ SFTP MIRROR ~~~")
	verboseprintf("%s\n", &creds)
	fmt.Printf("'%s' --> '%s'\n\n", local, remote)

	fmt.Println("reverse:", reverse, "| dry:", dry, "| ignorehidden", ignorehidden)

	var nItems, nBytes uint
	t0 := time.Now()

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
	verboseprintf("SFTP connection established; %s\n", &creds)

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

	// verboseprintf("local %v\n", filesetLocal.String())
	// verboseprintf("local %v\n", filesetRemote.String())

	basepath := strings.TrimSuffix(filesetLocal.Basepath, string(os.PathSeparator))

	// step 1: copy everything from source to dst if src newer
	err = filepath.Walk(local,
		func(srcPath string, srcInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			childPath := strings.TrimPrefix(srcPath, filesetLocal.Basepath)
			if childPath == basepath {
				return nil // skip basepath
			}

			if ignorehidden && (strings.HasPrefix(srcPath, ".") || strings.Contains(srcPath, "/.")) {
				verboseprintf("skip hidden '%s'\n", srcPath)
				return nil
			}

			nItems++
			nBytes += uint(srcInfo.Size())

			dstPath := filepath.Join(remote, childPath)

			// A) item is directory.
			//   exists in dst?
			//     no  --> create.
			//     yes --> skip.
			if srcInfo.IsDir() {
				verboseprintf("create or skip dir '%s'\n", dstPath)
				return sc.MkdirAll(dstPath)
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
			if !filesetRemote.Contains(childPath) {
				fmt.Printf("copy file '%s'\n", srcPath)
				if !dry {
					_, err := libsftp.UploadFile(sc, srcPath, dstPath)
					return err
				}
				return nil
			}

			dstInfo, _ := sc.Stat(filepath.Join(filesetRemote.Basepath, childPath))
			if compare.BasicUnequal(srcInfo, dstInfo) {
				fmt.Printf("overwrite file '%s'\n", srcPath)
				if dry {
					return nil
				}
				_, err := libsftp.UploadFile(sc, srcPath, dstPath)
				return err
			} else {
				verboseprintf("skip file '%s'\n", srcPath)
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	// step 2: clean everything from dst that is not in src
	if clean {
		// for file in filesetDst: file exists in filesetSrc ? --> Delete if not.
		for name, dstInfo := range filesetRemote.Paths {
			if skipHidden && (strings.HasPrefix(name, ".") || strings.Contains(name, "/.")) {
				verboseprintf("skip hidden '%s'\n", name)
				return nil
			}
			if !filesetLocal.Contains(name) {
				fmt.Printf("file/dir '%v' does not exist in src, delete\n", name)
				if dry {
					return nil
				}
				if dstInfo.IsDir() {
					err = sc.RemoveDirectory(filepath.Join(filesetRemote.Basepath, name))
				} else {
					// err = sc.Remove(filepath.Join(filesetRemote.Basepath, name))
					err = libsftp.DeleteFile(sc, filepath.Join(filesetRemote.Basepath, name), true)
				}
				if err != nil {
					// this can sometimes give an error if the parent directory was deleted before...
					verboseprint("deletion failed,", err)
				}
			}
		}
	}

	dt := time.Since(t0)
	verboseprintf("~~~ SFTP MIRROR done ~~~\n%v items (%v) in %v\n~~~\n",
		nItems,
		copy.ByteCount(nBytes),
		dt,
	)

	return nil
}
