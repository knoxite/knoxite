/*
 * knoxite
 *     Copyright (c) 2019, Fabian Siegel <fabians1999@gmail.com>
 *
 *   For license see LICENSE
 */

package sftp

import (
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	agent "golang.org/x/crypto/ssh/agent"
	kh "golang.org/x/crypto/ssh/knownhosts"

	knoxite "github.com/knoxite/knoxite/lib"
)

type StorageSFTP struct {
	url  url.URL
	ssh  *ssh.Client
	sftp *sftp.Client
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterBackendFactory(&StorageSFTP{})
}

func (*StorageSFTP) NewBackend(u url.URL) (knoxite.Backend, error) {
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
			return &StorageSFTP{}, knoxite.ErrInvalidPassword
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
		return &StorageSFTP{}, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return &StorageSFTP{}, err
	}

	storage := StorageSFTP{
		url:  u,
		sftp: client,
		ssh:  conn,
	}

	storagesftp, err := knoxite.NewStorageFilesystem(u.Path, &storage)
	storage.StorageFilesystem = storagesftp
	if err != nil {
		return &StorageSFTP{}, err
	}

	return &storage, nil
}

func (backend *StorageSFTP) Protocols() []string {
	return []string{"sftp"}
}

func (backend *StorageSFTP) AvailableSpace() (uint64, error) {
	stat, err := backend.sftp.StatVFS(backend.url.Path)
	return stat.FreeSpace(), err
}

func (backend *StorageSFTP) Close() error {
	return backend.sftp.Close()
}

func (backend *StorageSFTP) Description() string {
	return "SSH/SFTP Storage"
}

func (backend *StorageSFTP) Location() string {
	return backend.url.String()
}

func (backend *StorageSFTP) CreatePath(path string) error {
	return backend.sftp.MkdirAll(path)
}

func (backend *StorageSFTP) DeleteFile(path string) error {
	return backend.sftp.Remove(path)
}

func (backend *StorageSFTP) DeletePath(path string) error {
	files, err := backend.sftp.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			err = backend.DeletePath(backend.sftp.Join(path, file.Name()))
		} else {
			err = backend.DeleteFile(backend.sftp.Join(path, file.Name()))
		}
		if err != nil {
			return err
		}
	}
	return backend.sftp.Remove(path)
}

func (backend *StorageSFTP) ReadFile(path string) ([]byte, error) {
	file, err := backend.sftp.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func (backend *StorageSFTP) WriteFile(path string, data []byte) (size uint64, err error) {
	file, err := backend.sftp.Create(path)
	if err != nil {
		return 0, err
	}
	length, err := file.Write(data)
	return uint64(length), err
}

func (backend *StorageSFTP) Stat(path string) (uint64, error) {
	stat, err := backend.sftp.Stat(path)
	if err != nil {
		return 0, err
	}
	return uint64(stat.Size()), err
}
