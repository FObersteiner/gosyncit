package libsftp

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Credentials for SSH auth
type Credentials struct {
	Usr        string
	Host       string
	AgentSock  string
	Port       int
	SSHtimeout time.Duration
}

func (c *Credentials) String() string {
	return fmt.Sprintf("User: %v, on: %v:%v", c.Usr, c.Host, c.Port)
}

// GetSSHconn tries to establish an SSH connection with given Credentials
func GetSSHconn(creds Credentials) (*ssh.Client, error) {
	// access keyring to unlock SSH key later on.
	sock, err := net.Dial("unix", os.Getenv(creds.AgentSock))
	if err != nil {
		return nil, fmt.Errorf("unable to call SSH socket: %v", err)
	}
	defer sock.Close()

	agentsock := agent.NewClient(sock)

	// if the keyring is unlocked with user login, signer keys are also unlocked here:
	signers, err := agentsock.Signers()
	if err != nil {
		return nil, fmt.Errorf("create signers error: %s", err)
	}

	// get host key from known_hosts file
	hostKey, err := GetHostKey(creds.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain host key: %s", err)
	}

	// complete SSH config:
	sshConfig := &ssh.ClientConfig{
		User: creds.Usr,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
		// the default host key is ssh-rsa:
		HostKeyAlgorithms: []string{"ssh-rsa"},
		// HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		HostKeyCallback: ssh.FixedHostKey(hostKey),
		Timeout:         creds.SSHtimeout,
	}

	// Connect to server via SSH
	return ssh.Dial("tcp", fmt.Sprintf("%v:%v", creds.Host, creds.Port), sshConfig)
}

// GetHostKey from local known hosts
func GetHostKey(host string) (ssh.PublicKey, error) {
	var hk ssh.PublicKey
	home, err := os.UserHomeDir()
	if err != nil {
		return hk, err
	}
	file, err := os.Open(filepath.Join(home, ".ssh", "known_hosts"))
	if err != nil {
		return hk, fmt.Errorf("unable to read known_hosts file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hk, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return hk, fmt.Errorf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hk == nil {
		return hk, fmt.Errorf("no hostkey found for %s", host)
	}

	return hk, err
}

// ListDirsFiles in a certain directory "remoteDir" on the SFTP server
func ListDirsFiles(sc *sftp.Client, remoteDir string) (fsinfo []fs.FileInfo, err error) {
	fsinfo, err = sc.ReadDir(remoteDir)
	if err != nil {
		return fsinfo, fmt.Errorf("unable to list remote dir: %v", err)
	}

	return fsinfo, err
}

// UploadFile to SFTP server. The directory path on the remote must exist.
func UploadFile(sc *sftp.Client, localFile, remoteFile string) (n int64, err error) {
	srcFile, err := os.Open(localFile)
	if err != nil {
		return 0, fmt.Errorf("unable to open local file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := sc.OpenFile(remoteFile, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return 0, fmt.Errorf("unable to open remote file: %v", err)
	}
	defer dstFile.Close()

	n, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return 0, fmt.Errorf("unable to upload local file: %v", err)
	}

	// srcInfo, err := os.Stat(localFile)
	// if err != nil {
	// 	return n, fmt.Errorf("unable to get local file stats: %v", err)
	// }

	// mtime := srcInfo.ModTime()
	// err = sc.Chtimes(remoteFile, mtime, mtime)
	// if err != nil {
	// 	return n, fmt.Errorf("unable to set local file timestamp to remote file: %v", err)
	// }

	return n, nil
}

// DownloadFile from SFTP server
func DownloadFile(sc *sftp.Client, remoteFile, localFile string) (n int64, err error) {
	srcFile, err := sc.OpenFile(remoteFile, (os.O_RDONLY))
	if err != nil {
		return 0, fmt.Errorf("unable to open remote file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localFile)
	if err != nil {
		return 0, fmt.Errorf("unable to open local file: %v", err)
	}
	defer dstFile.Close()

	n, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return 0, fmt.Errorf("unable to download remote file: %v", err)
	}

	return n, err
}

// DeleteFile from SFTP server.
// A wrapper around sftp.Client.Remove and sftp.Client.RemoveDirectory.
// If removeDir is true but the directory is not empty, an error will be returned.
func DeleteFile(sc *sftp.Client, remoteFile string, removeDir bool) (err error) {
	err = sc.Remove(remoteFile)
	if err == nil && removeDir {
		dir := filepath.Dir(remoteFile)
		err = sc.RemoveDirectory(dir)
	}

	return err
}
