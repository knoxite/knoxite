/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package ftp

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"

	knoxite "github.com/knoxite/knoxite/lib"
)

// StorageFTP stores data on a remote FTP
type StorageFTP struct {
	url   url.URL
	ftp   *ftp.ServerConn
	login bool
	knoxite.StorageFilesystem
}

// Error declaration
var (
	ErrInvalidAuthentication = errors.New("Wrong Username or Password")
)

func init() {
	knoxite.RegisterBackendFactory(&StorageFTP{})
}

// NewBackend establishes an FTP connection and returns a StorageFTP backend
func (*StorageFTP) NewBackend(u url.URL) (knoxite.Backend, error) {
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil || len(port) == 0 {
		port = "ftp"
		u.Host = net.JoinHostPort(u.Host, port)
	}

	// Starting a connection
	con, err := ftp.DialTimeout(u.Host, 30*time.Second)
	if err != nil {
		return &StorageFTP{}, err
	}

	// Authenticate the client if desired
	loggedIn := false
	if u.User != nil && len(u.User.Username()) > 0 {
		// Doesn't matter if pw exists
		pw, _ := u.User.Password()
		err = con.Login(u.User.Username(), pw)
		if err != nil {
			return &StorageFTP{}, ErrInvalidAuthentication
		}
		loggedIn = true
	}

	storage := StorageFTP{
		url:   u,
		ftp:   con,
		login: loggedIn,
	}

	storageftp, err := knoxite.NewStorageFilesystem(u.Path, &storage)
	storage.StorageFilesystem = storageftp
	if err != nil {
		return &StorageFTP{}, err
	}

	return &storage, nil
}

// Location returns the type and location of the repository
func (backend *StorageFTP) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageFTP) Close() error {
	if backend.login {
		if err := backend.ftp.Logout(); err != nil {
			return err
		}
	}
	return backend.ftp.Quit()
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageFTP) Protocols() []string {
	return []string{"ftp"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageFTP) Description() string {
	return "FTP Storage"
}

// AvailableSpace returns the free space on this backen
func (backend *StorageFTP) AvailableSpace() (uint64, error) {
	return 0, knoxite.ErrAvailableSpaceUnknown
}

// CreatePath creates a dir including all its parent dirs, when required
func (backend *StorageFTP) CreatePath(path string) error {
	slicedPath := strings.Split(path, "/")
	for i := range slicedPath {
		if i == 0 {
			// don't try to create root-dir
			continue
		}
		if err := backend.ftp.MakeDir(filepath.Join(slicedPath[:i+1]...)); err != nil {
			// We only want to return an error when creating the last directory
			// in this path failed. Parent dirs _may_ already exist
			if i+1 == len(slicedPath) {
				return err
			}
		}
	}

	return nil
}

// Stat returns the size of a file on ftp
func (backend *StorageFTP) Stat(path string) (uint64, error) {
	base, last := filepath.Split(path)
	entries, err := backend.ftp.List(base)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, &os.PathError{Op: "stat", Path: path, Err: errors.New("error reading metadata")}
	}
	for _, v := range entries {
		if v.Name == last {
			return v.Size, nil
		}
	}

	return 0, &os.PathError{Op: "stat", Path: path, Err: errors.New("error reading metadata")}
}

// ReadFile reads a file from ftp
func (backend *StorageFTP) ReadFile(path string) ([]byte, error) {
	file, err := backend.ftp.Retr(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

// WriteFile writes file to ftp
func (backend *StorageFTP) WriteFile(path string, data []byte) (size uint64, err error) {
	err = backend.ftp.Stor(path, bytes.NewReader(data))
	return uint64(len(data)), err
}

// DeleteFile deletes a file from ftp
func (backend *StorageFTP) DeleteFile(path string) error {
	return backend.ftp.Delete(path)
}

// DeletePath deletes a directory including all its content from ftp
func (backend *StorageFTP) DeletePath(path string) error {
	list, err := backend.ftp.List(path)
	if err != nil {
		return err
	}
	for _, l := range list {
		if l.Type == ftp.EntryTypeFolder {
			if len(l.Name) == 0 || strings.HasPrefix(l.Name, ".") {
				continue
			}
			err = backend.DeletePath(filepath.Join(path, l.Name))
			if err != nil {
				return err
			}
		}
		if l.Type == ftp.EntryTypeFile {
			err = backend.ftp.Delete(filepath.Join(path, l.Name))
			if err != nil {
				return err
			}
		}
	}

	return backend.ftp.RemoveDir(path)
}
