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
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/kothar/go-backblaze.v0"
)

// StorageBackblaze stores data on a remote Backblaze
type StorageBackblaze struct {
	url            url.URL
	repositoryFile string
	chunkIndexFile string
	bucket         *backblaze.Bucket
	backblaze      *backblaze.B2
}

// NewStorageBackblaze returns a
func NewStorageBackblaze(URL url.URL) (*StorageBackblaze, error) {
	// Checking username and password
	if URL.User.Username() == "" {
		return &StorageBackblaze{}, ErrInvalidUsername
	}
	pw, pwexist := URL.User.Password()
	if !pwexist {
		return &StorageBackblaze{}, ErrInvalidPassword
	}

	// Creating a new Client for accessing the B2 API
	cl, err := backblaze.NewB2(backblaze.Credentials{
		AccountID:      URL.User.Username(),
		ApplicationKey: pw,
	})
	if err != nil {
		return &StorageBackblaze{}, err
	}

	// Creating the bucket prefixes
	bucketPrefix := strings.Split(URL.Path, "/")
	if len(bucketPrefix) != 2 {
		return &StorageBackblaze{}, ErrInvalidRepositoryURL
	}

	// Getting/Creating a bucket for backblaze
	var bucket *backblaze.Bucket
	bucket, err = cl.Bucket(bucketPrefix[1])
	if err != nil || bucket == nil {
		// Bucket probably doesn't exists
		bucket, err = cl.CreateBucket(bucketPrefix[1], backblaze.AllPrivate)
		if err != nil {
			// Bucket exists but we don't have access on it
			return &StorageBackblaze{}, err
		}
	}
	return &StorageBackblaze{
		url:            URL,
		repositoryFile: bucketPrefix[1] + "-repository",
		chunkIndexFile: bucketPrefix[1] + "-chunkindex",
		bucket:         bucket,
		backblaze:      cl,
	}, nil

}

// Location returns the type and location of the repository
func (backend *StorageBackblaze) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageBackblaze) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageBackblaze) Protocols() []string {
	return []string{"backblaze"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageBackblaze) Description() string {
	return "Backblaze Storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageBackblaze) AvailableSpace() (uint64, error) {
	// Currently not supported
	return 0, ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from backblaze
func (backend *StorageBackblaze) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)
	_, obj, err := backend.bucket.DownloadFileByName(fileName)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(obj)
	return &data, err
}

// StoreChunk stores a single Chunk on backblaze
func (backend *StorageBackblaze) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	buf := bytes.NewBuffer(*data)
	metadata := make(map[string]string)
	i, err := backend.bucket.UploadFile(fileName, metadata, buf)
	if err != nil {
		return 0, ErrStoreChunkFailed
	}
	file := backblaze.File(*i)
	return uint64(file.ContentLength), err
}

// DeleteChunk deletes a single Chunk
func (backend *StorageBackblaze) DeleteChunk(shasum string, parts, totalParts uint) error {
	// FIXME: implement this
	return ErrDeleteChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageBackblaze) LoadSnapshot(id string) ([]byte, error) {
	_, obj, err := backend.bucket.DownloadFileByName("snapshot-" + id)
	if err != nil {
		return nil, ErrSnapshotNotFound
	}
	return ioutil.ReadAll(obj)
}

// SaveSnapshot stores a snapshot
func (backend *StorageBackblaze) SaveSnapshot(id string, data []byte) error {
	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	_, err := backend.bucket.UploadFile("snapshot-"+id, metadata, buf)
	return err
}

// LoadChunkIndex reads the chunk-index
func (backend *StorageBackblaze) LoadChunkIndex() ([]byte, error) {
	_, obj, err := backend.bucket.DownloadFileByName(backend.chunkIndexFile)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveChunkIndex stores the chunk-index
func (backend *StorageBackblaze) SaveChunkIndex(data []byte) error {
	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	_, err := backend.bucket.UploadFile(backend.chunkIndexFile, metadata, buf)
	return err
}

// InitRepository creates a new repository
func (backend *StorageBackblaze) InitRepository() error {
	var placeholder []byte
	buf := bytes.NewBuffer(placeholder)

	// Creating the files on backblaze
	metadata := make(map[string]string)

	if _, err := backend.bucket.UploadFile(backend.repositoryFile, metadata, buf); err != nil {
		return err
	}
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageBackblaze) LoadRepository() ([]byte, error) {
	_, obj, err := backend.bucket.DownloadFileByName(backend.repositoryFile)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveRepository stores the metadata for a repository
func (backend *StorageBackblaze) SaveRepository(data []byte) error {
	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	_, err := backend.bucket.UploadFile(backend.repositoryFile, metadata, buf)
	return err
}
