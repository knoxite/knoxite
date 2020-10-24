/*
 * knoxite
 *     Copyright (c) 2016, Johannes Fürmann <fuermannj+floss@gmail.com>
 *
 *   For license see LICENSE
 */

package amazons3

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/knoxite/knoxite"
)

func init() {
	knoxite.RegisterStorageBackend(&AmazonS3StorageBackend{})
}

// NewBackend initializes an Amazon S3 Storage Backend.
func (*AmazonS3StorageBackend) NewBackend(url url.URL) (knoxite.Backend, error) {
	// Set up Session for Amazon S3 Client
	sessionConfig := aws.NewConfig()
	if region := url.Query().Get("region"); region != "" {
		sessionConfig.Region = aws.String(region)
	}

	if endpoint := url.Query().Get("endpoint"); endpoint != "" {
		sessionConfig.Endpoint = aws.String(endpoint)
	}

	if forcePathStyle := url.Query().Get("force_path_style"); forcePathStyle == "true" {
		sessionConfig.S3ForcePathStyle = aws.Bool(true)
	}

	sesn, err := session.NewSession(sessionConfig)
	if err != nil {
		return &AmazonS3StorageBackend{}, err
	}

	new := &AmazonS3StorageBackend{
		url:        url,
		service:    s3.New(sesn),
		bucketName: url.Hostname(),
	}

	fs, err := knoxite.NewStorageFilesystem(url.Path, new)
	if err != nil {
		return &AmazonS3StorageBackend{}, err
	}

	new.StorageFilesystem = fs

	return new, nil
}

// Protocols returns a list of supported Protocol Handlers.
func (*AmazonS3StorageBackend) Protocols() []string {
	return []string{"amazons3"}
}
