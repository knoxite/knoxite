/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/knoxite/knoxite"
)

// HTTPStorage stores data on a remote HTTP server.
type HTTPStorage struct {
	URL url.URL
}

func init() {
	knoxite.RegisterStorageBackend(&HTTPStorage{})
}

// NewBackend returns a HTTPStorage backend.
func (*HTTPStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	return &HTTPStorage{
		URL: u,
	}, nil
}

// Location returns the type and location of the repository.
func (backend *HTTPStorage) Location() string {
	return backend.URL.String()
}

// Close the backend.
func (backend *HTTPStorage) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *HTTPStorage) Protocols() []string {
	return []string{"http", "https"}
}

// Description returns a user-friendly description for this backend.
func (backend *HTTPStorage) Description() string {
	return "HTTP(S) Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *HTTPStorage) AvailableSpace() (uint64, error) {
	return uint64(0), knoxite.ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from network.
func (backend *HTTPStorage) LoadChunk(shasum string, part, totalParts uint) ([]byte, error) {
	//	fmt.Printf("Fetching from: %s.\n", backend.URL+"/download/"+chunk.ShaSum)
	res, err := http.Get(backend.URL.String() + "/download/" + shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10))
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return []byte{}, knoxite.ErrLoadChunkFailed
	}

	return ioutil.ReadAll(res.Body)
}

// StoreChunk stores a single Chunk on network.
func (backend *HTTPStorage) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, werr := bodyWriter.CreateFormFile("uploadfile", shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))
	if werr != nil {
		fmt.Println("error writing to buffer")
		return 0, werr
	}

	_, err = fileWriter.Write(data)
	if err != nil {
		return 0, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(backend.URL.String()+"/upload", contentType, bodyBuf)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, knoxite.ErrStoreChunkFailed
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	//	fmt.Printf("\tUploaded chunk: %d bytes\n", len(*data))
	return uint64(len(data)), err
}

// DeleteChunk deletes a single Chunk.
func (backend *HTTPStorage) DeleteChunk(shasum string, parts, totalParts uint) error {
	// FIXME: implement this
	return knoxite.ErrDeleteChunkFailed
}

// LoadSnapshot loads a snapshot.
func (backend *HTTPStorage) LoadSnapshot(id string) ([]byte, error) {
	//	fmt.Printf("Fetching snapshot from: %s.\n", backend.URL+"/snapshot/"+id)
	res, err := http.Get(backend.URL.String() + "/snapshot/" + id)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

// SaveSnapshot stores a snapshot.
func (backend *HTTPStorage) SaveSnapshot(id string, data []byte) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", id)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	_, err = fileWriter.Write(data)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(backend.URL.String()+"/snapshot", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return knoxite.ErrStoreSnapshotFailed
	}
	//	fmt.Printf("Uploaded snapshot: %d bytes\n", len(data))
	return err
}

// LoadChunkIndex reads the chunk-index.
func (backend *HTTPStorage) LoadChunkIndex() ([]byte, error) {
	//	fmt.Printf("Fetching chunk-index from: %s.\n", backend.URL+"/chunkindex")
	res, err := http.Get(backend.URL.String() + "/chunkindex")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

// SaveChunkIndex stores the chunk-index.
func (backend *HTTPStorage) SaveChunkIndex(data []byte) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", "chunkindex")
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	_, err = fileWriter.Write(data)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(backend.URL.String()+"/chunkindex", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return knoxite.ErrStoreChunkIndexFailed
	}
	//	fmt.Printf("Uploaded chunk-index: %d bytes\n", len(data))
	return err
}

// InitRepository creates a new repository.
func (backend *HTTPStorage) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository.
func (backend *HTTPStorage) LoadRepository() ([]byte, error) {
	//	fmt.Printf("Fetching repository from: %s.\n", backend.URL+"/repository")
	res, err := http.Get(backend.URL.String() + "/repository")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

// SaveRepository stores the metadata for a repository.
func (backend *HTTPStorage) SaveRepository(data []byte) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", "repository.knoxite")
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	_, err = fileWriter.Write(data)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(backend.URL.String()+"/repository", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return knoxite.ErrStoreRepositoryFailed
	}
	//	fmt.Printf("Uploaded repository: %d bytes\n", len(data))
	return err
}

// LockRepository locks the repository and prevents other instances from
// concurrent access.
func (backend *HTTPStorage) LockRepository(data []byte) ([]byte, error) {
	// TODO: implement
	return nil, nil
}

// UnlockRepository releases the lock.
func (backend *HTTPStorage) UnlockRepository() error {
	// TODO: implement
	return nil
}
