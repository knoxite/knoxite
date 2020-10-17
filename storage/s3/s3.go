package s3

import (
	"bytes"
	"io"
	"net/url"
	"path"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/knoxite/knoxite"
)

// S3Storage stores data on a remote AmazonS3.
type S3Storage struct {
	Client       *awsS3.S3
	URL          url.URL
	BucketName   string
	BucketPrefix string
}

func init() {
	knoxite.RegisterStorageBackend(&S3Storage{})
}

// NewBackend returns a S3Storage backend.
func (*S3Storage) NewBackend(URL url.URL) (knoxite.Backend, error) {
	sessionConfig := &aws.Config{}

	// Set AWS region if supplied in query
	if region := URL.Query().Get("region"); region != "" {
		sessionConfig.Region = &region
	}

	// Set AWS Endpoint URL if supplied in query
	if endpointURL := URL.Query().Get("endpoint"); endpointURL != "" {
		sessionConfig.Endpoint = &endpointURL
	}

	new := &S3Storage{
		Client:       awsS3.New(awsSession.New(sessionConfig)),
		BucketName:   URL.Host,
		BucketPrefix: URL.Path,
		URL:          URL,
	}

	return new, nil
}

// Location returns the type and location of the repository.
func (backend *S3Storage) Location() string {
	return backend.URL.String()
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
	return backend.s3Load(backend.chunkPath(shasum, part, totalParts))
}

// StoreChunk stores a single Chunk on network.
func (backend *S3Storage) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	chunkKey := backend.chunkPath(shasum, part, totalParts)

	// Check if chunk already exists, return 0 otherwise.
	_, err = backend.Client.HeadObject(&awsS3.HeadObjectInput{
		Bucket: &backend.BucketName,
		Key:    &chunkKey,
	})

	// Here, we expect either no error at all (which means the chunk already
	// exists) or an erorr with Error Code `NotFound` or `NoSuchKey`, which means
	// the chunk doesn't exist yet and we have to write it.
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// An unexpected error happened during our HEAD request, which
			// means we shouldn't continue.
			switch awsErr.Code() {
			case "NotFound":
			case "NoSuchKey":
			default:
				return 0, err
			}
		}
	} else {
		// chunk already exists
		return 0, nil
	}

	return uint64(len(data)), backend.s3Store(data, chunkKey)
}

// DeleteChunk deletes a single Chunk.
func (backend *S3Storage) DeleteChunk(shasum string, part, totalParts uint) error {
	key := backend.chunkPath(shasum, part, totalParts)

	_, err := backend.Client.DeleteObject(&awsS3.DeleteObjectInput{
		Bucket: &backend.BucketName,
		Key:    &key,
	})

	return err
}

// LoadSnapshot loads a snapshot.
func (backend *S3Storage) LoadSnapshot(id string) ([]byte, error) {
	key := path.Join(backend.BucketPrefix, "snapshots", id)
	return backend.s3Load(key)
}

// SaveSnapshot stores a snapshot.
func (backend *S3Storage) SaveSnapshot(id string, data []byte) error {
	key := path.Join(backend.BucketPrefix, "snapshots", id)
	return backend.s3Store(data, key)
}

// LoadChunkIndex reads the chunk-index.
func (backend *S3Storage) LoadChunkIndex() ([]byte, error) {
	key := path.Join(backend.BucketPrefix, "chunks", "index")
	return backend.s3Load(key)
}

// SaveChunkIndex stores the chunk-index.
func (backend *S3Storage) SaveChunkIndex(data []byte) error {
	key := path.Join(backend.BucketPrefix, "chunks", "index")
	return backend.s3Store(data, key)
}

// InitRepository creates a new repository.
func (backend *S3Storage) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository.
func (backend *S3Storage) LoadRepository() ([]byte, error) {
	key := path.Join(backend.BucketPrefix, knoxite.RepoFilename)
	return backend.s3Load(key)
}

// SaveRepository stores the metadata for a repository.
func (backend *S3Storage) SaveRepository(data []byte) error {
	key := path.Join(backend.BucketPrefix, knoxite.RepoFilename)
	return backend.s3Store(data, key)
}

func (backend *S3Storage) s3Store(data []byte, key string) error {
	_, err := backend.Client.PutObject(
		&awsS3.PutObjectInput{
			Body:   bytes.NewReader(data),
			Bucket: &backend.BucketName,
			Key:    &key,
		},
	)

	// spew.Dump(out)

	return err
}

func (backend *S3Storage) s3Load(key string) ([]byte, error) {
	out, err := backend.Client.GetObject(
		&awsS3.GetObjectInput{
			Bucket: &backend.BucketName,
			Key:    &key,
		},
	)

	if err != nil {
		return []byte{}, err
	}

	data := make([]byte, *out.ContentLength)
	if _, err := out.Body.Read(data); err != io.EOF {
		return data, err
	}

	// spew.Dump(data)

	return data, err
}

func (backend *S3Storage) chunkPath(shasum string, part, totalParts uint) string {
	return path.Join(
		backend.BucketPrefix, "chunks", shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10),
	)
}
