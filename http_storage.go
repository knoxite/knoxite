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
)

// HTTPStorage stores data on a remote HTTP server
type HTTPStorage struct {
	URL string
}

// Location returns the type and location of the repository
func (backend *HTTPStorage) Location() string {
	return ""
}

// Close the backend
func (backend *HTTPStorage) Close() error {
	return nil
}

// Protocol Scheme supported by this backend
func (backend *HTTPStorage) Protocol() string {
	return "http"
}

// Description returns a user-friendly description for this backend
func (backend *HTTPStorage) Description() string {
	return "HTTP(S) Storage"
}

// LoadChunk loads a Chunk from network
func (backend *HTTPStorage) LoadChunk(chunk Chunk) ([]byte, error) {
	//	fmt.Printf("Fetching from: %s.\n", backend.URL+"/download/"+chunk.ShaSum)
	res, err := http.Get(backend.URL + "/download/" + chunk.ShaSum)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	//	fmt.Printf("Download finished: %d bytes\n", len(b))
	return b, err
}

// StoreChunk stores a single Chunk on network
func (backend *HTTPStorage) StoreChunk(chunk Chunk, data *[]byte) (size uint64, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", chunk.ShaSum)
	if err != nil {
		fmt.Println("error writing to buffer")
		return 0, err
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
		return 0, errors.New("Storing chunk failed")
	}
	//	fmt.Printf("\tUploaded chunk: %d bytes\n", len(*data))
	return uint64(len(*data)), err
}

// LoadSnapshot loads a snapshot
func (backend *HTTPStorage) LoadSnapshot(id string) ([]byte, error) {
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
		return errors.New("Storing snapshot failed")
	}
	//	fmt.Printf("Uploaded snapshot: %d bytes\n", len(data))
	return err
}

// InitRepository creates a new repository
func (backend *HTTPStorage) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *HTTPStorage) LoadRepository() ([]byte, error) {
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
func (backend *HTTPStorage) SaveRepository(data []byte) error {
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
		return errors.New("Storing repository failed")
	}
	//	fmt.Printf("Uploaded repository: %d bytes\n", len(data))
	return err
}
