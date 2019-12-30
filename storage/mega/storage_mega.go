/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package mega

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/t3rm1n4l/go-mega"

	knoxite "github.com/knoxite/knoxite/lib"
)

// StorageMega stores data on a remote Mega
type StorageMega struct {
	url  url.URL
	mega *mega.Mega
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterBackendFactory(&StorageMega{})
}

// NewBackend returns a StorageMega backend
func (*StorageMega) NewBackend(u url.URL) (knoxite.Backend, error) {
	backend := StorageMega{
		url:  u,
		mega: mega.New(),
	}

	// checking for username and password
	if u.User == nil || u.User.Username() == "" {
		return &StorageMega{}, knoxite.ErrInvalidUsername
	}
	pw, pwexist := u.User.Password()
	if !pwexist {
		return &StorageMega{}, knoxite.ErrInvalidPassword
	}

	err := backend.mega.Login(u.User.Username(), pw)
	// log into the mega client for accessing the API
	if err != nil {
		return &StorageMega{}, err
	}

	knfs, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &StorageMega{}, err
	}

	backend.StorageFilesystem = knfs
	return &backend, nil
}

// Location returns the type and location of the repository
func (backend *StorageMega) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageMega) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageMega) Protocols() []string {
	return []string{"mega"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageMega) Description() string {
	return "mega.nz storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageMega) AvailableSpace() (uint64, error) {
	quota, err := backend.mega.GetQuota()
	if err != nil {
		return 0, knoxite.ErrAvailableSpaceUnknown
	}

	return quota.Mstrg - quota.Cstrg, nil
}

// CreatePath creates a dir including all its parent dirs, when required
func (backend *StorageMega) CreatePath(path string) error {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	slicedPath := strings.Split(path, "/")
	currentRoot := backend.mega.FS.GetRoot()

	for _, pathSlice := range slicedPath {
		// get all nodes in current root directory
		nodesInCurrentRoot, err := backend.mega.FS.PathLookup(currentRoot, []string{pathSlice})
		if err != nil {
			// if we couldn't find a node at this path we need to create it
			currentRoot, err = backend.mega.CreateDir(pathSlice, currentRoot)
			if err != nil {
				return err
			}
		} else {
			// if nodes with the same name exist we take the node at index 0
			currentRoot = nodesInCurrentRoot[0]
		}
	}
	return nil
}

// Stat returns the size of a file
func (backend *StorageMega) Stat(path string) (uint64, error) {
	node, err := backend.getNodeFromPath(path)
	if err != nil {
		return 0, err
	}

	return uint64(node.GetSize()), nil
}

// ReadFile reads a file from mega
func (backend *StorageMega) ReadFile(path string) ([]byte, error) {
	nodeToRead, err := backend.getNodeFromPath(path)
	if err != nil {
		return nil, err
	}

	download, err := backend.mega.NewDownload(&nodeToRead)
	if err != nil {
		return nil, err
	}

	var bytes []byte
	for i := 0; i < download.Chunks(); i++ {
		chunkBytes, err := download.DownloadChunk(i)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, chunkBytes...)
	}

	download.Finish()
	return bytes, nil
}

// WriteFile write files on mega
func (backend *StorageMega) WriteFile(path string, data []byte) (size uint64, err error) {
	dir, file := filepath.Split(path)

	_, err = backend.getNodeFromPath(path)
	if err == nil {
		// sadly, if the file exists it needs to be deleted before re-uploading, otherwise there will be a copy
		err = backend.DeleteFile(path)
		if err != nil {
			return 0, err
		}
	}

	nodeToWriteIn, err := backend.getNodeFromPath(dir)
	if err != nil {
		return 0, err
	}

	upload, err := backend.mega.NewUpload(&nodeToWriteIn, file, int64(len(data)))
	if err != nil {
		return 0, err
	}

	for id := 0; id < upload.Chunks(); id++ {
		chk_start, chk_size, err := upload.ChunkLocation(id)
		if err != nil {
			return 0, err
		}
		err = upload.UploadChunk(id, data[chk_start:chk_start+int64(chk_size)])
		if err != nil {
			return 0, err
		}
	}
	_, err = upload.Finish()
	return uint64(len(data)), err
}

// DeleteFile deletes a file from mega
func (backend *StorageMega) DeleteFile(path string) error {
	fileToDelete, err := backend.getNodeFromPath(path)
	if err != nil {
		return err
	}

	err = backend.mega.Delete(&fileToDelete, true)
	if err != nil {
		return err
	}

	return err
}

// getNodeFromPath() returns the last node in a path on mega. It may be a file or a directory node.
func (backend *StorageMega) getNodeFromPath(path string) (mega.Node, error) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	slicedPath := strings.Split(path, "/")

	// initially get mega filesystem root node to start our lookup from
	currentRoot := backend.mega.FS.GetRoot()
	for i, pathSlice := range slicedPath {
		// get all nodes in current root directory
		nodesInCurrentRoot, err := backend.mega.FS.PathLookup(currentRoot, []string{pathSlice})
		if err != nil {
			return mega.Node{}, err
		}

		// finding folder node by pathSlice
		found := false
		for _, node := range nodesInCurrentRoot {
			if node.GetName() == pathSlice {
				currentRoot = node
				found = true
				break
			}
		}
		if !found {
			return mega.Node{}, errors.New("file or directory not found on mega: " + pathSlice)
		}
		// last element of slicedPath is the actual file/directory node
		if i == len(slicedPath)-1 {
			return *currentRoot, nil
		}
	}
	return mega.Node{}, errors.New("file or directory not found on mega")
}
