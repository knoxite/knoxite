/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
 *
 *   For license see LICENSE
 */

package s3

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/minio/minio-go"

	"github.com/knoxite/knoxite"
)

// S3Storage stores data on a remote AmazonS3.
type S3Storage struct {
	url              url.URL
	chunkBucket      string
	snapshotBucket   string
	repositoryBucket string
	region           string
	client           *minio.Client
}

func init() {
	knoxite.RegisterStorageBackend(&S3Storage{})
}

// NewBackend returns a S3Storage backend.
func (*S3Storage) NewBackend(URL url.URL) (knoxite.Backend, error) {
	var ssl bool
	switch URL.Scheme {
	case "s3":
		ssl = false
	case "s3s":
		ssl = true
	default:
		return &S3Storage{}, errors.New("Invalid s3 url scheme")
	}

	var username, pw string
	if URL.User != nil {
		username = URL.User.Username()
		pw, _ = URL.User.Password()
	}
	if len(username) == 0 {
		username = os.Getenv("AWS_ACCESS_KEY_ID")
		if len(username) == 0 {
			return &S3Storage{}, knoxite.ErrInvalidUsername
		}
	}
	if len(pw) == 0 {
		pw = os.Getenv("AWS_SECRET_ACCESS_KEY")
		if len(pw) == 0 {
			return &S3Storage{}, knoxite.ErrInvalidPassword
		}
	}

	regionAndBucketPrefix := strings.Split(URL.Path, "/")
	if len(regionAndBucketPrefix) != 3 {
		return &S3Storage{}, knoxite.ErrInvalidRepositoryURL
	}

	cl, err := minio.New(URL.Host, username, pw, ssl)
	if err != nil {
		return &S3Storage{}, err
	}

	return &S3Storage{url: URL,
		client:           cl,
		region:           regionAndBucketPrefix[1],
		chunkBucket:      regionAndBucketPrefix[2] + "-chunks",
		snapshotBucket:   regionAndBucketPrefix[2] + "-snapshots",
		repositoryBucket: regionAndBucketPrefix[2] + "-repository",
	}, nil
}

// Location returns the type and location of the repository.
func (backend *S3Storage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *S3Storage) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *S3Storage) Protocols() []string {
	return []string{"s3", "s3s"}
}

// Description returns a user-friendly description for this backend.
func (backend *S3Storage) Description() string {
	return "Amazon S3 Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *S3Storage) AvailableSpace() (uint64, error) {
	return uint64(0), knoxite.ErrAvailableSpaceUnlimited
}

// LoadChunk loads a Chunk from network.
func (backend *S3Storage) LoadChunk(shasum string, part, totalParts uint) ([]byte, error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)
	obj, err := backend.client.GetObject(backend.chunkBucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// StoreChunk stores a single Chunk on network.
func (backend *S3Storage) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	if _, err = backend.client.StatObject(backend.chunkBucket, fileName, minio.StatObjectOptions{}); err == nil {
		// Chunk is already stored
		return 0, nil
	}

	buf := bytes.NewBuffer(data)
	i, err := backend.client.PutObject(backend.chunkBucket, fileName, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return uint64(i), err
}

// DeleteChunk deletes a single Chunk.
func (backend *S3Storage) DeleteChunk(shasum string, part, totalParts uint) error {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	err := backend.client.RemoveObject(backend.chunkBucket, fileName)
	if err != nil {
		return err
	}

	return nil
}

// LoadSnapshot loads a snapshot.
func (backend *S3Storage) LoadSnapshot(id string) ([]byte, error) {
	obj, err := backend.client.GetObject(backend.snapshotBucket, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// SaveSnapshot stores a snapshot.
func (backend *S3Storage) SaveSnapshot(id string, data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := backend.client.PutObject(backend.snapshotBucket, id, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

// LoadChunkIndex reads the chunk-index.
func (backend *S3Storage) LoadChunkIndex() ([]byte, error) {
	obj, err := backend.client.GetObject(backend.chunkBucket, knoxite.ChunkIndexFilename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// SaveChunkIndex stores the chunk-index.
func (backend *S3Storage) SaveChunkIndex(data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := backend.client.PutObject(backend.chunkBucket, knoxite.ChunkIndexFilename, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

// InitRepository creates a new repository.
func (backend *S3Storage) InitRepository() error {
	chunkBucketExist, err := backend.client.BucketExists(backend.chunkBucket)
	if err != nil {
		return err
	}
	if !chunkBucketExist {
		err = backend.client.MakeBucket(backend.chunkBucket, backend.region)
		if err != nil {
			return err
		}
	} else {
		return knoxite.ErrRepositoryExists
	}

	snapshotBucketExist, err := backend.client.BucketExists(backend.snapshotBucket)
	if err != nil {
		return err
	}
	if !snapshotBucketExist {
		err = backend.client.MakeBucket(backend.snapshotBucket, backend.region)
		if err != nil {
			return err
		}
	} else {
		return knoxite.ErrRepositoryExists
	}

	repositoryBucketExist, err := backend.client.BucketExists(backend.repositoryBucket)
	if err != nil {
		return err
	}
	if !repositoryBucketExist {
		err = backend.client.MakeBucket(backend.repositoryBucket, backend.region)
		if err != nil {
			return err
		}
	} else {
		return knoxite.ErrRepositoryExists
	}

	return nil
}

// LoadRepository reads the metadata for a repository.
func (backend *S3Storage) LoadRepository() ([]byte, error) {
	obj, err := backend.client.GetObject(backend.repositoryBucket, knoxite.RepoFilename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ioutil.ReadAll(obj)
}

// SaveRepository stores the metadata for a repository.
func (backend *S3Storage) SaveRepository(data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := backend.client.PutObject(backend.repositoryBucket, knoxite.RepoFilename, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

// LockRepository locks the repository and prevents other instances from
// concurrent access.
func (backend *S3Storage) LockRepository(data []byte) ([]byte, error) {
	// TODO: implement
	return nil, nil
}

// UnlockRepository releases the lock.
func (backend *S3Storage) UnlockRepository() error {
	// TODO: implement
	return nil
}
