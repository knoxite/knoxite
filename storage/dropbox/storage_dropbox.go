/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package dropbox

import (
	"bytes"
	"io/ioutil"
	"net/url"

	"github.com/tj/go-dropbox"
	"github.com/tj/go-dropy"

	knoxite "github.com/knoxite/knoxite/lib"
)

// StorageDropbox stores data on a remote Dropbox
type StorageDropbox struct {
	url   url.URL
	dropy *dropy.Client
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterBackendFactory(&StorageDropbox{})
}

// NewBackend returns a StorageDropbox backend
func (*StorageDropbox) NewBackend(u url.URL) (knoxite.Backend, error) {
	pw, pwexist := u.User.Password()
	if !pwexist {
		return &StorageDropbox{}, knoxite.ErrInvalidPassword
	}

	backend := StorageDropbox{
		url:   u,
		dropy: dropy.New(dropbox.New(dropbox.NewConfig(pw))),
	}

	storageDB, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &StorageDropbox{}, err
	}
	backend.StorageFilesystem = storageDB

	return &backend, nil
}

// Location returns the type and location of the repository
func (backend *StorageDropbox) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageDropbox) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageDropbox) Protocols() []string {
	return []string{"dropbox"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageDropbox) Description() string {
	return "Dropbox Storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageDropbox) AvailableSpace() (uint64, error) {
	space, err := backend.dropy.Client.Users.GetSpaceUsage()
	if err != nil {
		return 0, err
	}
	return space.Allocation.Allocated, nil
}

// CreatePath creates a dir including all its parent dirs, when required
func (backend *StorageDropbox) CreatePath(path string) error {
	return backend.dropy.Mkdir(path)
}

// Stat returns the size of a file
func (backend *StorageDropbox) Stat(path string) (uint64, error) {
	fileinfo, err := backend.dropy.Stat(path)
	if err != nil {
		return 0, err
	}
	return uint64(fileinfo.Size()), nil
}

// ReadFile reads a file from dropbox
func (backend *StorageDropbox) ReadFile(path string) ([]byte, error) {
	file, err := backend.dropy.Download(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

// WriteFile write files on dropbox
func (backend *StorageDropbox) WriteFile(path string, data []byte) (size uint64, err error) {
	return uint64(len(data)), backend.dropy.Upload(path, bytes.NewReader(data))
}

// DeleteFile deletes a file from dropbox
func (backend *StorageDropbox) DeleteFile(path string) error {
	return backend.dropy.Delete(path)
}
