/*
 * knoxite
 *     Copyright (c) 2020, Johannes Fürmann <fuermannj+floss@gmail.com>
 *
 *   For license see LICENSE
 */

package amazons3

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/knoxite/knoxite"
)

// Stat returns the size of the object with key `path` if successful and 0 as
// well as an error otherwise.
func (backend *AmazonS3StorageBackend) Stat(path string) (uint64, error) {
	out, err := backend.service.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(backend.bucketName),
		Key:    aws.String(path),
	})

	if err != nil {
		return 0, err
	}

	return uint64(*out.ContentLength), nil
}

// CreatePath creates a folder in a filesystem-like storage backend.
func (*AmazonS3StorageBackend) CreatePath(path string) error {
	// In S3, this is a no-op since "paths" are just a convention and
	// automatically created once you write an object.
	return nil
}

// ReadFile reads a file from the backend.
func (backend *AmazonS3StorageBackend) ReadFile(path string) ([]byte, error) {
	result, err := backend.service.GetObject(&s3.GetObjectInput{
		Key:    aws.String(path),
		Bucket: aws.String(backend.bucketName),
	})
	if err != nil {
		return nil, err
	}

	resultBytes, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err != nil {
		return nil, err
	}

	return resultBytes, nil
}

// WriteFile writes a file to the storage backend.
func (backend *AmazonS3StorageBackend) WriteFile(path string, data []byte) (uint64, error) {
	databuf := bytes.NewReader(data)

	_, err := backend.service.PutObject(&s3.PutObjectInput{
		Key:    aws.String(path),
		Bucket: aws.String(backend.bucketName),
		Body:   databuf,
	})

	if err != nil {
		return 0, err
	}

	// Since "Content-Length" is not part of the PutObject method's response
	// in the S3 API, we kind of cheat here by telling knoxite what it wants
	// to hear instead of verifying.
	// The only way we could find that out at this point is via an additional
	// HEAD request, increasing latency and cost.

	return uint64(len(data)), nil
}

// DeleteFile deletes a file from the storage backend.
func (backend *AmazonS3StorageBackend) DeleteFile(path string) error {
	_, err := backend.service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(backend.bucketName),
		Key:    aws.String(path),
	})

	return err
}

// Close closes the StorageFileSystem.
func (*AmazonS3StorageBackend) Close() error {
	// Close is meaningless for S3 since it's using a RESTful API which is
	// stateless by nature and doesn't need closing.
	return nil
}

// Description returns a human-readable description for the Storage backend.
func (*AmazonS3StorageBackend) Description() string {
	return "Amazon S3 Storage Backend (using AWS SDK)"
}

// Location returns the backend's URL as a string.
func (backend *AmazonS3StorageBackend) Location() string {
	return backend.url.String()
}

// AvailableSpace returns the free space on this backend.
func (*AmazonS3StorageBackend) AvailableSpace() (uint64, error) {
	// Amazon S3 doesn't constrain bucket size, so we treat it as unlimited storage
	return 0, knoxite.ErrAvailableSpaceUnlimited
}
