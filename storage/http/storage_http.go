/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
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

	knoxite "github.com/knoxite/knoxite/lib"
)

// StorageHTTP stores data on a remote HTTP server
type StorageHTTP struct {
	URL url.URL
}

func init() {
	knoxite.RegisterBackendFactory(&StorageHTTP{})
}

// NewBackend returns a StorageHTTP backend
func (*StorageHTTP) NewBackend(u url.URL) (knoxite.Backend, error) {
	return &StorageHTTP{
		URL: u,
	}, nil
}

// Location returns the type and location of the repository
func (backend *StorageHTTP) Location() string {
	return backend.URL.String()
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
	return uint64(0), knoxite.ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from network
func (backend *StorageHTTP) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	//	fmt.Printf("Fetching from: %s.\n", backend.URL+"/download/"+chunk.ShaSum)
	res, err := http.Get(backend.URL.String() + "/download/" + shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10))
	if err != nil {
		return &[]byte{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return &[]byte{}, knoxite.ErrLoadChunkFailed
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

	resp, err := http.Post(backend.URL.String()+"/upload", contentType, bodyBuf)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, knoxite.ErrStoreChunkFailed
	}

	//	fmt.Printf("\tUploaded chunk: %d bytes\n", len(*data))
	return uint64(len(*data)), err
}

// DeleteChunk deletes a single Chunk
func (backend *StorageHTTP) DeleteChunk(shasum string, parts, totalParts uint) error {
	// FIXME: implement this
	return knoxite.ErrDeleteChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageHTTP) LoadSnapshot(id string) ([]byte, error) {
	//	fmt.Printf("Fetching snapshot from: %s.\n", backend.URL+"/snapshot/"+id)
	res, err := http.Get(backend.URL.String() + "/snapshot/" + id)
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

// LoadChunkIndex reads the chunk-index
func (backend *StorageHTTP) LoadChunkIndex() ([]byte, error) {
	//	fmt.Printf("Fetching chunk-index from: %s.\n", backend.URL+"/chunkindex")
	res, err := http.Get(backend.URL.String() + "/chunkindex")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	//	fmt.Printf("Downloading rchunk-index finished: %d bytes\n", len(b))
	return b, err
}

// SaveChunkIndex stores the chunk-index
func (backend *StorageHTTP) SaveChunkIndex(data []byte) error {
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

// InitRepository creates a new repository
func (backend *StorageHTTP) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageHTTP) LoadRepository() ([]byte, error) {
	//	fmt.Printf("Fetching repository from: %s.\n", backend.URL+"/repository")
	res, err := http.Get(backend.URL.String() + "/repository")
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
