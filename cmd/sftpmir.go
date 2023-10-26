/*
Copyright © 2023 Florian Obersteiner <f.obersteiner@posteo.de>

License: see LICENSE in the root directory of the repo.
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

// SftpMir mirrors directory 'local' to 'remote' (SFTP) or vice versa (see 'reverse' flag).
func SftpMir(local, remote string, creds libsftp.Credentials, reverse, dry, ignorehidden, clean bool) error {
	fmt.Println("~~~ SFTP MIRROR ~~~")
	verboseprintf("%s\n", &creds)
	arrow := "-->"
	if reverse {
		arrow = "<--"
	}
	fmt.Printf("'%s' %v '%s'\n\n", local, arrow, remote)

	if reverse {
		return sftpRemoteToLocal(local, remote, creds, dry, ignorehidden, clean)
	}
	return sftpLocalToRemote(local, remote, creds, dry, ignorehidden, clean)
}

// wrapper for default mirror case; local to remote
func sftpLocalToRemote(local, remote string, creds libsftp.Credentials, dry, ignorehidden, clean bool) error {
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
	basepath := strings.TrimSuffix(filesetLocal.Basepath, string(os.PathSeparator))

	// step 1: copy everything from local to remote if src newer (or size different)
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

	// step 2: clean everything from remote that is not in local
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

// wrapper if reverse == True; mirror from remote to local
func sftpRemoteToLocal(local, remote string, creds libsftp.Credentials, dry, ignorehidden, clean bool) error {
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

	basepath := strings.TrimSuffix(filesetLocal.Basepath, string(os.PathSeparator))

	// step 1: copy everything from remote to local if newer
	walker := sc.Walk(filesetRemote.Basepath)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}

		srcPath := walker.Path()

		childPath := strings.TrimPrefix(srcPath, filesetRemote.Basepath)

		if childPath == "" {
			continue
		}

		if childPath == basepath {
			continue
		}

		if ignorehidden && (strings.HasPrefix(srcPath, ".") || strings.Contains(srcPath, "/.")) {
			verboseprintf("skip hidden '%s'\n", srcPath)
			return nil
		}

		srcInfo := walker.Stat()

		nItems++
		nBytes += uint(srcInfo.Size())

		dstPath := filepath.Join(local, childPath)

		// A) item is directory.
		//   exists in dst?
		//     no  --> create.
		//     yes --> skip.
		if srcInfo.IsDir() { // dir is created locally
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
		if !filesetLocal.Contains(childPath) {
			fmt.Printf("copy file '%s'\n", srcPath)
			if !dry {
				_, err := libsftp.DownloadFile(sc, srcPath, dstPath)
				return err
			}
			return nil
		}

		dstInfo, _ := os.Stat(filepath.Join(filesetLocal.Basepath, childPath))
		if compare.BasicUnequal(srcInfo, dstInfo) {
			fmt.Printf("overwrite file '%s'\n", srcPath)
			// fmt.Println(srcInfo.ModTime(), dstInfo.ModTime())
			// fmt.Println(srcInfo.Size(), dstInfo.Size())
			if dry {
				return nil
			}
			_, err := libsftp.DownloadFile(sc, srcPath, dstPath)
			if err != nil {
				return err
			}
			// // SFTP timestamps tend to be messed up; make sure it is set:
			// mtime := srcInfo.ModTime()
			// err = os.Chtimes(dstPath, mtime, mtime)
			return err
		} else {
			verboseprintf("skip file '%s'\n", srcPath)
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
