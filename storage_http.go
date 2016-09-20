/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
)

// Error declarations
var (
	ErrChunkNotFound         = errors.New("Loading chunk failed")
	ErrStoreChunkFailed      = errors.New("Storing chunk failed")
	ErrStoreSnapshotFailed   = errors.New("Storing snapshot failed")
	ErrStoreRepositoryFailed = errors.New("Storing repository failed")
)

// StorageHTTP stores data on a remote HTTP server
type StorageHTTP struct {
	URL string
}

// Location returns the type and location of the repository
func (backend *StorageHTTP) Location() string {
	return backend.URL
}

// Close the backend
func (backend *StorageHTTP) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageHTTP) Protocols() []string {
	return []string{"http", "https"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageHTTP) Description() string {
	return "HTTP(S) Storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageHTTP) AvailableSpace() (uint64, error) {
	return uint64(0), ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from network
func (backend *StorageHTTP) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	//	fmt.Printf("Fetching from: %s.\n", backend.URL+"/download/"+chunk.ShaSum)
	res, err := http.Get(backend.URL + "/download/" + shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return &[]byte{}, ErrChunkNotFound
	}

	b, err := ioutil.ReadAll(res.Body)
	return &b, err
}

// StoreChunk stores a single Chunk on network
func (backend *StorageHTTP) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, werr := bodyWriter.CreateFormFile("uploadfile", shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))
	if werr != nil {
		fmt.Println("error writing to buffer")
		return 0, werr
	}

	_, err = fileWriter.Write(*data)
	if err != nil {
		return 0, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(backend.URL+"/upload", contentType, bodyBuf)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, ErrStoreChunkFailed
	}

	//	fmt.Printf("\tUploaded chunk: %d bytes\n", len(*data))
	return uint64(len(*data)), err
}

// LoadSnapshot loads a snapshot
func (backend *StorageHTTP) LoadSnapshot(id string) ([]byte, error) {
	//	fmt.Printf("Fetching snapshot from: %s.\n", backend.URL+"/snapshot/"+id)
	res, err := http.Get(backend.URL + "/snapshot/" + id)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	//	fmt.Printf("Downloading snapshot finished: %d bytes\n", len(b))
	return b, err
}

// SaveSnapshot stores a snapshot
func (backend *StorageHTTP) SaveSnapshot(id string, data []byte) error {
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

	resp, err := http.Post(backend.URL+"/snapshot", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return ErrStoreSnapshotFailed
	}
	//	fmt.Printf("Uploaded snapshot: %d bytes\n", len(data))
	return err
}

// InitRepository creates a new repository
func (backend *StorageHTTP) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageHTTP) LoadRepository() ([]byte, error) {
	//	fmt.Printf("Fetching repository from: %s.\n", backend.URL+"/repository")
	res, err := http.Get(backend.URL + "/repository")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	//	fmt.Printf("Downloading repository finished: %d bytes\n", len(b))
	return b, err
}

// SaveRepository stores the metadata for a repository
func (backend *StorageHTTP) SaveRepository(data []byte) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", "repository.knox")
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

	resp, err := http.Post(backend.URL+"/repository", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return ErrStoreRepositoryFailed
	}
	//	fmt.Printf("Uploaded repository: %d bytes\n", len(data))
	return err
}
