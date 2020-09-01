/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package backblaze

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/kothar/go-backblaze.v0"

	"github.com/knoxite/knoxite"
)

// BackblazeStorage stores data on a remote Backblaze.
type BackblazeStorage struct {
	url            url.URL
	repositoryFile string
	chunkIndexFile string
	Bucket         *backblaze.Bucket
	backblaze      *backblaze.B2
}

func init() {
	knoxite.RegisterStorageBackend(&BackblazeStorage{})
}

// NewBackend returns a BackblazeStorage backend.
func (*BackblazeStorage) NewBackend(URL url.URL) (knoxite.Backend, error) {
	// Checking username and password
	if URL.User == nil || URL.User.Username() == "" {
		return &BackblazeStorage{}, knoxite.ErrInvalidUsername
	}
	pw, pwexist := URL.User.Password()
	if !pwexist {
		return &BackblazeStorage{}, knoxite.ErrInvalidPassword
	}

	// Creating a new Client for accessing the B2 API
	cl, err := backblaze.NewB2(backblaze.Credentials{
		KeyID:          URL.User.Username(),
		ApplicationKey: pw,
	})
	if err != nil {
		return &BackblazeStorage{}, err
	}

	// Creating the Bucket prefixes
	bucketPrefix := strings.Split(URL.Path, "/")
	if len(bucketPrefix) != 2 {
		return &BackblazeStorage{}, knoxite.ErrInvalidRepositoryURL
	}

	// Getting/Creating a Bucket for backblaze
	var bucket *backblaze.Bucket
	bucket, err = cl.Bucket(bucketPrefix[1])
	if err != nil || bucket == nil {
		// Bucket probably doesn't exist
		bucket, err = cl.CreateBucket(bucketPrefix[1], backblaze.AllPrivate)
		if err != nil {
			// Bucket exists but we don't have access to it
			return &BackblazeStorage{}, err
		}
	}
	return &BackblazeStorage{
		url:            URL,
		repositoryFile: bucketPrefix[1] + "-repository",
		chunkIndexFile: bucketPrefix[1] + "-chunkindex",
		Bucket:         bucket,
		backblaze:      cl,
	}, nil
}

// Location returns the type and location of the repository.
func (backend *BackblazeStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *BackblazeStorage) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *BackblazeStorage) Protocols() []string {
	return []string{"backblaze"}
}

// Description returns a user-friendly description for this backend.
func (backend *BackblazeStorage) Description() string {
	return "Backblaze Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *BackblazeStorage) AvailableSpace() (uint64, error) {
	// Currently not supported
	return 0, knoxite.ErrAvailableSpaceUnlimited
}

// LoadChunk loads a Chunk from backblaze.
func (backend *BackblazeStorage) LoadChunk(shasum string, part, totalParts uint) ([]byte, error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)
	_, obj, err := backend.Bucket.DownloadFileByName(fileName)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// StoreChunk stores a single Chunk on backblaze.
func (backend *BackblazeStorage) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	files, err := backend.findLatestFileVersion(fileName)
	if err == nil && len(files) > 0 {
		if files[0].Size == len(data) {
			return 0, nil
		}
	}

	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	file, err := backend.upload(fileName, metadata, buf)
	if err != nil {
		return 0, err
	}
	return uint64(file.ContentLength), nil
}

// DeleteChunk deletes a single Chunk.
func (backend *BackblazeStorage) DeleteChunk(shasum string, part, totalParts uint) error {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	files, err := backend.findLatestFileVersion(fileName)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return knoxite.ErrDeleteChunkFailed
	}

	_, err = backend.Bucket.DeleteFileVersion(fileName, files[0].ID)
	return err
}

// LoadSnapshot loads a snapshot.
func (backend *BackblazeStorage) LoadSnapshot(id string) ([]byte, error) {
	_, obj, err := backend.Bucket.DownloadFileByName("snapshot-" + id)
	if err != nil {
		return nil, knoxite.ErrSnapshotNotFound
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// SaveSnapshot stores a snapshot.
func (backend *BackblazeStorage) SaveSnapshot(id string, data []byte) error {
	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	_, err := backend.upload("snapshot-"+id, metadata, buf)
	return err
}

// LoadChunkIndex reads the chunk-index.
func (backend *BackblazeStorage) LoadChunkIndex() ([]byte, error) {
	_, obj, err := backend.Bucket.DownloadFileByName(backend.chunkIndexFile)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// SaveChunkIndex stores the chunk-index.
func (backend *BackblazeStorage) SaveChunkIndex(data []byte) error {
	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	_, err := backend.upload(backend.chunkIndexFile, metadata, buf)
	return err
}

// InitRepository creates a new repository.
func (backend *BackblazeStorage) InitRepository() error {
	var placeholder []byte
	buf := bytes.NewBuffer(placeholder)

	// Creating the files on backblaze
	metadata := make(map[string]string)

	if _, err := backend.upload(backend.repositoryFile, metadata, buf); err != nil {
		return err
	}
	return nil
}

// LoadRepository reads the metadata for a repository.
func (backend *BackblazeStorage) LoadRepository() ([]byte, error) {
	files, err := backend.findLatestFileVersion(backend.repositoryFile)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, knoxite.ErrLoadRepositoryFailed
	}

	_, obj, err := backend.backblaze.DownloadFileByID(files[0].ID)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// SaveRepository stores the metadata for a repository.
func (backend *BackblazeStorage) SaveRepository(data []byte) error {
	buf := bytes.NewBuffer(data)
	metadata := make(map[string]string)
	_, err := backend.upload(backend.repositoryFile, metadata, buf)
	return err
}

func (backend *BackblazeStorage) findLatestFileVersion(fileName string) ([]backblaze.FileStatus, error) {
	var files []backblaze.FileStatus

	list, err := backend.Bucket.ListFileVersions(fileName, "", 1)
	if err != nil {
		return files, err
	}

	for _, v := range list.Files {
		if v.Name != fileName {
			continue
		}

		files = append(files, v)
	}

	return files, nil
}

func (backend *BackblazeStorage) upload(name string, meta map[string]string, file io.Reader) (*backblaze.File, error) {
	// delete existing versions of a file, before reuploading
	files, err := backend.findLatestFileVersion(backend.repositoryFile)
	if err != nil {
		return nil, err
	}

	for _, v := range files {
		_, err := backend.Bucket.DeleteFileVersion(v.Name, v.ID)
		if err != nil {
			return nil, err
		}
	}

	return backend.Bucket.UploadFile(name, meta, file)
}
