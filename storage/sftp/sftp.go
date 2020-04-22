/*
 * knoxite
 *     Copyright (c) 2019, Fabian Siegel <fabians1999@gmail.com>
 *                   2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package sftp

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	kh "golang.org/x/crypto/ssh/knownhosts"

	"github.com/knoxite/knoxite"
)

type SFTPStorage struct {
	url  url.URL
	ssh  *ssh.Client
	sftp *sftp.Client
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterStorageBackend(&SFTPStorage{})
}

func (*SFTPStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil || len(port) == 0 {
		port = "22"
		u.Host = net.JoinHostPort(u.Host, port)
	}
	username := u.User.Username()
	password, isSet := u.User.Password()

	auth := []ssh.AuthMethod{}
	if isSet {
		auth = append(auth, ssh.Password(password))
	} else {
		socket := os.Getenv("SSH_AUTH_SOCK")
		agent_conn, err := net.Dial("unix", socket)
		if err != nil {
			return &SFTPStorage{}, knoxite.ErrInvalidPassword
		} else {
			agentClient := agent.NewClient(agent_conn)
			auth = append(auth, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	usr, _ := user.Current()

	hostKeyCallback, err := kh.New(filepath.Join(usr.HomeDir, ".ssh/known_hosts"))
	if err != nil {
		// If no hostkey can be found, ignore it for now...
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
	}

	conn, err := ssh.Dial("tcp", u.Hostname()+":"+u.Port(), config)
	if err != nil {
		return &SFTPStorage{}, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return &SFTPStorage{}, err
	}

	backend := SFTPStorage{
		url:  u,
		sftp: client,
		ssh:  conn,
	}

	fs, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &SFTPStorage{}, err
	}
	backend.StorageFilesystem = fs

	return &backend, nil
}

func (backend *SFTPStorage) Protocols() []string {
	return []string{"sftp"}
}

func (backend *SFTPStorage) AvailableSpace() (uint64, error) {
	stat, err := backend.sftp.StatVFS(backend.url.Path)
	if err != nil || stat == nil {
		return 0, knoxite.ErrAvailableSpaceUnknown
	}

	return stat.FreeSpace(), err
}

func (backend *SFTPStorage) Close() error {
	return backend.sftp.Close()
}

func (backend *SFTPStorage) Description() string {
	return "SSH/SFTP Storage"
}

func (backend *SFTPStorage) Location() string {
	return backend.url.String()
}

func (backend *SFTPStorage) CreatePath(path string) error {
	return backend.sftp.MkdirAll(path)
}

func (backend *SFTPStorage) DeleteFile(path string) error {
	return backend.sftp.Remove(path)
}

func (backend *SFTPStorage) DeletePath(path string) error {
	fmt.Println("Deleting path", path)
	files, err := backend.sftp.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		fpath := backend.sftp.Join(path, file.Name())
		if file.IsDir() {
			err = backend.DeletePath(fpath)
			if err != nil {
				return err
			}
			err = backend.sftp.Remove(fpath)
		} else {
			err = backend.DeleteFile(fpath)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (backend *SFTPStorage) ReadFile(path string) ([]byte, error) {
	file, err := backend.sftp.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func (backend *SFTPStorage) WriteFile(path string, data []byte) (size uint64, err error) {
	file, err := backend.sftp.Create(path)
	if err != nil {
		return 0, err
	}
	length, err := file.Write(data)
	return uint64(length), err
}

func (backend *SFTPStorage) Stat(path string) (uint64, error) {
	stat, err := backend.sftp.Stat(path)
	if err != nil {
		return 0, err
	}
	return uint64(stat.Size()), err
}
