/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package ftp

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"

	"github.com/knoxite/knoxite"
)

// FTPStorage stores data on a remote FTP.
type FTPStorage struct {
	url   url.URL
	ftp   *ftp.ServerConn
	login bool
	knoxite.StorageFilesystem
}

// Error declarations.
var (
	ErrInvalidAuthentication = errors.New("wrong Username or Password")
)

func init() {
	knoxite.RegisterStorageBackend(&FTPStorage{})
}

// NewBackend establishes an FTP connection and returns a FTPStorage backend.
func (*FTPStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil || len(port) == 0 {
		port = "ftp"
		u.Host = net.JoinHostPort(u.Host, port)
	}

	// Starting a connection
	con, err := ftp.DialTimeout(u.Host, 30*time.Second)
	if err != nil {
		return &FTPStorage{}, err
	}

	// Authenticate the client if desired
	loggedIn := false
	if u.User != nil && len(u.User.Username()) > 0 {
		// Doesn't matter if pw exists
		pw, _ := u.User.Password()
		err = con.Login(u.User.Username(), pw)
		if err != nil {
			return &FTPStorage{}, ErrInvalidAuthentication
		}
		loggedIn = true
	}

	backend := FTPStorage{
		url:   u,
		ftp:   con,
		login: loggedIn,
	}

	fs, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &FTPStorage{}, err
	}
	backend.StorageFilesystem = fs

	return &backend, nil
}

// Location returns the type and location of the repository.
func (backend *FTPStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *FTPStorage) Close() error {
	if backend.login {
		if err := backend.ftp.Logout(); err != nil {
			return err
		}
	}
	return backend.ftp.Quit()
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *FTPStorage) Protocols() []string {
	return []string{"ftp"}
}

// Description returns a user-friendly description for this backend.
func (backend *FTPStorage) Description() string {
	return "FTP Storage"
}

// AvailableSpace returns the free space on this backen.
func (backend *FTPStorage) AvailableSpace() (uint64, error) {
	return 0, knoxite.ErrAvailableSpaceUnknown
}

// CreatePath creates a dir including all its parent dirs, when required.
func (backend *FTPStorage) CreatePath(path string) error {
	slicedPath := strings.Split(path, "/")
	for i := range slicedPath {
		if i == 0 {
			// don't try to create root-dir
			continue
		}
		_ = backend.ftp.MakeDir(filepath.Join(slicedPath[:i+1]...))
	}

	return nil
}

// Stat returns the size of a file on ftp.
func (backend *FTPStorage) Stat(path string) (uint64, error) {
	size, err := backend.ftp.FileSize(path)
	return uint64(size), err
}

// ReadFile reads a file from ftp.
func (backend *FTPStorage) ReadFile(path string) ([]byte, error) {
	file, err := backend.ftp.Retr(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

// WriteFile writes file to ftp.
func (backend *FTPStorage) WriteFile(path string, data []byte) (size uint64, err error) {
	err = backend.ftp.Stor(path, bytes.NewReader(data))
	return uint64(len(data)), err
}

// DeleteFile deletes a file from ftp.
func (backend *FTPStorage) DeleteFile(path string) error {
	return backend.ftp.Delete(path)
}

// DeletePath deletes a directory including all its content from ftp.
func (backend *FTPStorage) DeletePath(path string) error {
	fmt.Println("Deleting path", path)
	list, err := backend.ftp.List("")
	if err != nil {
		return err
	}
	for _, l := range list {
		fpath := filepath.Join(path, l.Name)

		if l.Type == ftp.EntryTypeFolder {
			if len(l.Name) == 0 ||
				strings.HasPrefix(l.Name, ".") ||
				l.Name == "virtual" {
				continue
			}

			err = backend.ftp.ChangeDir(l.Name)
			if err != nil {
				return err
			}
			err = backend.DeletePath(fpath)
			if err != nil {
				return err
			}
			err = backend.ftp.ChangeDirToParent()
			if err != nil {
				return err
			}
			err = backend.ftp.RemoveDir(fpath)
			if err != nil {
				return err
			}
		}

		if l.Type == ftp.EntryTypeFile {
			err = backend.ftp.Delete(fpath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
