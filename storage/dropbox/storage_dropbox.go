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
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/stacktic/dropbox"

	knoxite "github.com/knoxite/knoxite/lib"
)

// StorageDropbox stores data on a remote Dropbox
type StorageDropbox struct {
	url url.URL
	db  *dropbox.Dropbox
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterBackendFactory(&StorageDropbox{})
}

// NewBackend returns a StorageDropbox backend
func (*StorageDropbox) NewBackend(u url.URL) (knoxite.Backend, error) {
	backend := StorageDropbox{
		url: u,
		db:  dropbox.NewDropbox(),
	}

	storageDB, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &StorageDropbox{}, err
	}
	backend.StorageFilesystem = storageDB

	ak, _ := base64.StdEncoding.DecodeString("aXF1bGs0a25vajIydGtt")
	as, _ := base64.StdEncoding.DecodeString("N3htbmlhcDV0cmE5NTE5")
	backend.db.SetAppInfo(string(ak), string(as))

	if backend.url.User == nil || len(backend.url.User.Username()) == 0 {
		if err := backend.db.Auth(); err != nil {
			return &StorageDropbox{}, err
		}
		backend.url.User = url.User(backend.db.AccessToken())
	} else {
		backend.db.SetAccessToken(backend.url.User.Username())
	}

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
	account, err := backend.db.GetAccountInfo()
	if err != nil {
		return 0, err
	}

	return uint64(account.QuotaInfo.Quota - account.QuotaInfo.Shared - account.QuotaInfo.Normal), nil
}

// CreatePath creates a dir including all its parent dirs, when required
func (backend *StorageDropbox) CreatePath(path string) error {
	slicedPath := strings.Split(path, string(filepath.Separator))
	for i := range slicedPath {
		if i == 0 {
			// don't try to create root-dir
			continue
		}
		if _, err := backend.db.CreateFolder(filepath.Join(slicedPath[:i+1]...)); err != nil {
			// We only want to return an error when creating the last directory
			// in this path failed. Parent dirs _may_ already exist
			if i+1 == len(slicedPath) {
				return err
			}
		}
	}
	return nil
}

// Stat returns the size of a file
func (backend *StorageDropbox) Stat(path string) (uint64, error) {
	entry, err := backend.db.Metadata(path, false, false, "", "", 1)
	if entry.IsDeleted {
		return 0, &os.PathError{Op: "stat", Path: path, Err: errors.New("error reading metadata")}
	}
	return uint64(entry.Bytes), err
}

// ReadFile reads a file from dropbox
func (backend *StorageDropbox) ReadFile(path string) ([]byte, error) {
	obj, _, err := backend.db.Download(path, "", 0)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// WriteFile write files on dropbox
func (backend *StorageDropbox) WriteFile(path string, data []byte) (size uint64, err error) {
	_, err = backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(data)), len(data), path, true, "")
	return uint64(len(data)), err
}

// DeleteFile deletes a file from dropbox
func (backend *StorageDropbox) DeleteFile(path string) error {
	_, err := backend.db.Delete(path)
	return err
}
