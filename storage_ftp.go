/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// StorageFTP stores data on a remote FTP
type StorageFTP struct {
	url   url.URL
	ftp   *ftp.ServerConn
	login bool
	StorageFilesystem
}

// Error declaration
var (
	ErrInvalidAuthentication = errors.New("Wrong Username or Password")
)

// NewStorageFTP establishs a FTP connection and returns a StorageFTP object.
func NewStorageFTP(u url.URL) (*StorageFTP, error) {
	// Starting a connection
	con, err := ftp.DialTimeout(u.Host, 30*time.Second)
	if err != nil {
		return &StorageFTP{}, err
	}

	// Authenticate the client if desired
	loggedIn := false
	if len(u.User.Username()) > 0 {
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
	storageftp, err := NewStorageFilesystem(u.Path, &storage)
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
	return 0, ErrAvailableSpaceUnknown
}

// CreatePath creates a dir including all its parent dirs, when required
func (backend *StorageFTP) CreatePath(path string) error {
	slicedPath := strings.Split(path, string(filepath.Separator))
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
		//FIXME: there's probably a proper error for this already
		return 0, errors.New("Couldn't stat path on FTP server")
	}
	for _, v := range entries {
		if v.Name == last {
			return v.Size, nil
		}
	}

	//FIXME: there's probably a proper error for this already
	return 0, errors.New("Not found")
}

// ReadFile reads a file from ftp
func (backend *StorageFTP) ReadFile(path string) (*[]byte, error) {
	file, err := backend.ftp.Retr(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	return &data, err
}

// WriteFile writes file to ftp
func (backend *StorageFTP) WriteFile(path string, data *[]byte) (size uint64, err error) {
	err = backend.ftp.Stor(path, bytes.NewReader(*data))
	return uint64(len(*data)), err
}
