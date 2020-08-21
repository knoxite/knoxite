/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *                   2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package googlecloud

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/knoxite/knoxite"
)

// GoogleCloudStorage stores data in a Google Cloud Storage bucket.
type GoogleCloudStorage struct {
	knoxite.StorageFilesystem
	url    url.URL
	client storage.Client
	bucket storage.BucketHandle
}

func init() {
	knoxite.RegisterStorageBackend(&GoogleCloudStorage{})
}

// NewBackend returns a GoogleCloudStorage backend.
// To create a storage client we need either the path to a credential JSON file set via the environment variable GOOGLE_APPLICATION_CREDENTIALS="[PATH]"
// or the path to the JSON file passed via the user parameter of the URL scheme.
func (*GoogleCloudStorage) NewBackend(URL url.URL) (knoxite.Backend, error) {
	var credentialsPath string

	// check if path to the credentials file was passed via URL scheme
	if URL.User == nil || URL.User.Username() == "" {
		// if not, we check if the path to the credentials file is provided via environment variable
		var credentialsSet bool
		credentialsPath, credentialsSet = os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
		if !credentialsSet {
			return &GoogleCloudStorage{}, errors.New("No valid JSON credentials file provided.")
		}
	} else {
		credentialsPath = URL.User.Username()
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return &GoogleCloudStorage{}, err
	}

	slicedPath := strings.Split(URL.Path, "/")
	var folderPath string
	if len(slicedPath) > 2 {
		folderPath = strings.Join(slicedPath[2:], "/")
	} else {
		return &GoogleCloudStorage{}, knoxite.ErrInvalidRepositoryURL
	}

	// we can have a bucket handle even if the bucket doesn't exist yet, so we check if we can access bucket attributes
	bucket := client.Bucket(slicedPath[1])
	_, err = bucket.Attrs(ctx)
	if err != nil {
		return &GoogleCloudStorage{}, err
	}

	// we can have an object handle even if the object doesn't exist yet, so we check if we can access object attributes
	folder := bucket.Object(folderPath)
	_, err = folder.Attrs(ctx)
	if err != nil {
		return &GoogleCloudStorage{}, err
	}

	backend := GoogleCloudStorage{
		url:    URL,
		client: *client,
		bucket: *bucket,
	}

	fs, err := knoxite.NewStorageFilesystem(folderPath, &backend)
	if err != nil {
		return &GoogleCloudStorage{}, err
	}
	backend.StorageFilesystem = fs

	return &backend, nil
}

// Location returns the type and location of the repository.
func (backend *GoogleCloudStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *GoogleCloudStorage) Close() error {
	return backend.client.Close()
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *GoogleCloudStorage) Protocols() []string {
	return []string{"googlecloudstorage"}
}

// Description returns a user-friendly description for this backend.
func (backend *GoogleCloudStorage) Description() string {
	return "Google Cloud Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *GoogleCloudStorage) AvailableSpace() (uint64, error) {
	// since google cloud storage doesn't have quota and you can store as much data as you want we return 0
	return 0, nil
}

// CreatePath is not needed in Google Cloud Storage backend bacause paths are automatically created when writing a file.
func (backend *GoogleCloudStorage) CreatePath(path string) error {
	return nil
}

// Stat returns the size of a file.
func (backend *GoogleCloudStorage) Stat(path string) (uint64, error) {
	folder := backend.bucket.Object(path)
	attrs, err := folder.Attrs(context.Background())
	if err != nil {
		return 0, err
	}

	return uint64(attrs.Size), nil
}

// ReadFile reads a file from Google Cloud Storage.
func (backend *GoogleCloudStorage) ReadFile(path string) ([]byte, error) {
	reader, err := backend.bucket.Object(path).NewReader(context.Background())
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	// read may return nil in some error situation so we need to check the error from close
	err = reader.Close()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// WriteFile writes a file on Google Cloud Storage.
func (backend *GoogleCloudStorage) WriteFile(path string, data []byte) (size uint64, err error) {
	writer := backend.bucket.Object(path).NewWriter(context.Background())
	// we set the ChunkSize to 0 to upload the data in a single request
	writer.ChunkSize = 0
	written, err := writer.Write(data)
	if err != nil {
		return 0, err
	}
	// write may return nil in some error situation so we need to check the error from close
	err = writer.Close()
	if err != nil {
		return 0, err
	}

	return uint64(written), nil
}

// DeleteFile deletes a file from Google Cloud Storage.
func (backend *GoogleCloudStorage) DeleteFile(path string) error {
	err := backend.bucket.Object(path).Delete(context.Background())
	if err != nil {
		return err
	}
	return nil
}
