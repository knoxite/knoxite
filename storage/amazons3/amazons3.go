/*
 * knoxite
 *     Copyright (c) 2020, Johannes Fürmann <fuermannj+floss@gmail.com>
 *
 *   For license see LICENSE
 */

package amazons3

import (
	"net/url"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/knoxite/knoxite"
)

// AmazonS3Client is an abstraction over the functions used from
// aws-golang-sdk's S3 client. Using this instead of the client
// directly makes it easier to mock.
type AmazonS3Client interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

// AmazonS3StorageBackend is the storage backend that adapts knoxite's backend
// interface to Amazon S3.
type AmazonS3StorageBackend struct {
	knoxite.StorageFilesystem

	url        url.URL
	service    AmazonS3Client
	bucketName string
}
