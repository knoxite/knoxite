/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/minio/minio-go"
)

// StorageAmazonS3 stores data on a remote AmazonS3
type StorageAmazonS3 struct {
	url          url.URL
	bucketPrefix string
	region       string
	client       *minio.Client
}

// NewStorageAmazonS3 returns a StorageAmazonS3 object.
func NewStorageAmazonS3(URL url.URL) (*StorageAmazonS3, error) {

	ssl := true
	switch URL.Scheme {
	case "s3":
		ssl = false
	case "s3s":
		ssl = true
	default:
		panic("Invalid s3 url scheme")
	}

	regionAndBucketPrefix := strings.Split(URL.Path, "/")
	if len(regionAndBucketPrefix) != 3 {
		return &StorageAmazonS3{}, ErrInvalidRepositoryURL
	}

	pw, _ := URL.User.Password()
	cl, err := minio.New(URL.Host, URL.User.Username(), pw, ssl)
	if err != nil {
		return &StorageAmazonS3{}, err
	}

	return &StorageAmazonS3{url: URL,
		client:       cl,
		region:       regionAndBucketPrefix[1],
		bucketPrefix: regionAndBucketPrefix[2],
	}, nil
}

// Location returns the type and location of the repository
func (backend *StorageAmazonS3) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageAmazonS3) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageAmazonS3) Protocols() []string {
	return []string{"s3", "s3s"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageAmazonS3) Description() string {
	return "Amazon S3 Storage"
}

// LoadChunk loads a Chunk from network
func (backend *StorageAmazonS3) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	return &[]byte{}, ErrChunkNotFound
}

// StoreChunk stores a single Chunk on network
func (backend *StorageAmazonS3) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	return 0, ErrStoreChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageAmazonS3) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot
func (backend *StorageAmazonS3) SaveSnapshot(id string, data []byte) error {
	return ErrStoreSnapshotFailed
}

// InitRepository creates a new repository
func (backend *StorageAmazonS3) InitRepository() error {
	chunkBucketExist, err := backend.client.BucketExists(backend.bucketPrefix + "-chunks")
	if err != nil {
		return err
	}
	if !chunkBucketExist {
		err = backend.client.MakeBucket(backend.bucketPrefix+"-chunks", backend.region)
		if err != nil {
			return err
		}
	} else {
		return ErrRepositoryExists
	}

	snapshotBucketExist, err := backend.client.BucketExists(backend.bucketPrefix + "-snapshots")
	if err != nil {
		return err
	}
	if !snapshotBucketExist {
		err = backend.client.MakeBucket(backend.bucketPrefix+"-snapshots", backend.region)
		if err != nil {
			return err
		}
	} else {
		return ErrRepositoryExists
	}

	repositoryBucketExist, err := backend.client.BucketExists(backend.bucketPrefix + "-repository")
	if err != nil {
		return err
	}
	if !repositoryBucketExist {
		err = backend.client.MakeBucket(backend.bucketPrefix+"-repository", backend.region)
		if err != nil {
			return err
		}
	} else {
		return ErrRepositoryExists
	}

	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageAmazonS3) LoadRepository() ([]byte, error) {
	obj, err := backend.client.GetObject(backend.bucketPrefix+"-repository", repoFilename)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveRepository stores the metadata for a repository
func (backend *StorageAmazonS3) SaveRepository(data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := backend.client.PutObject(backend.bucketPrefix+"-repository", repoFilename, buf, "application/octet-stream")
	return err
}
