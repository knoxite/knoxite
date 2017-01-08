// +build ci

/*
 * knoxite
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package dropbox

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	mrand "math/rand"
	"net/url"
	"os"
	"testing"
)

var (
	dropboxURL *url.URL
)

func TestMain(m *testing.M) {
	flag.Parse()

	path := os.Getenv("KNOXITE_DROPBOX_URL")
	if len(path) == 0 {
		panic("KNOXITE_DROPBOX_URL is undefined")
	}
	// create a random bucket name every time to avoid collisions
	rnd := make([]byte, 8)
	rand.Read(rnd)
	path += hex.EncodeToString(rnd)

	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}

	dropboxURL = u
	os.Exit(m.Run())
}

func teardownRepo(u *url.URL) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	db := b.(*StorageDropbox)
	err := db.DeleteFile(db.Path)
	if err != nil {
		panic(err)
	}
}

func TestStorageDropboxNewBackend(t *testing.T) {
	_, err := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err != nil {
		t.Error(err)
	}
}

func TestStorageDropboxLocation(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if b.Location() != dropboxURL.String() {
		t.Errorf("Expected %v, got %v", dropboxURL.String(), b.Location())
	}
}

func TestStorageDropboxProtocols(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if len(b.Protocols()) != 1 {
		t.Error("Invalid amount of protocols")
	}

	if b.Protocols()[0] != "dropbox" {
		t.Error("Invalid protocols")
	}
}

func TestStorageDropboxDescription(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)

	if b.Description() != "Dropbox Storage" {
		t.Error("Invalid Description")
	}
}

func TestStorageDropboxInitRepository(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)

	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)
}

func TestStorageDropboxSaveRepository(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnd := make([]byte, 256)
	rand.Read(rnd)

	err := b.SaveRepository(rnd)
	if err != nil {
		t.Error(err)
	}

	data, err := b.LoadRepository()
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnd) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnd[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageDropboxLoadRepository(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnd := make([]byte, 256)
	rand.Read(rnd)

	err := b.SaveRepository(rnd)
	if err != nil {
		t.Error(err)
	}

	data, err := b.LoadRepository()
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnd) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnd[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageDropboxSaveSnapshot(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	rndid := make([]byte, 8)
	rand.Read(rndid)
	id := hex.EncodeToString(rndid)

	err := b.SaveSnapshot(id, rnddata)
	if err != nil {
		t.Error(err)
	}

	data, err := b.LoadSnapshot(id)
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnddata) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageDropboxLoadSnapshot(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	rndid := make([]byte, 8)
	rand.Read(rndid)
	id := hex.EncodeToString(rndid)

	err := b.SaveSnapshot(id, rnddata)
	if err != nil {
		t.Error(err)
	}

	data, err := b.LoadSnapshot(id)
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnddata) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageDropboxStoreChunk(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Int())
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	shasumdata := sha256.Sum256(rnddata)
	shasum := hex.EncodeToString(shasumdata[:])

	size, err := b.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}
	if size != uint64(len(rnddata)) {
		t.Error("Data length missmatch")
	}

	// Test to store the same chunk twice. Size should be 0
	size, err = b.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}
	if size != 0 {
		t.Error("Already exisiting chunks should not be overwritten")
	}

	data, err := b.LoadChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}
	if len(*data) != len(rnddata) {
		t.Error("Data length missmatch")
	}
	for i := 0; i != len(*data); i++ {
		if (*data)[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageDropboxLoadChunk(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Int())
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	shasumdata := sha256.Sum256(rnddata)
	shasum := hex.EncodeToString(shasumdata[:])

	_, err := b.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}

	data, err := b.LoadChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}
	if len(*data) != len(rnddata) {
		t.Error("Data length missmatch")
	}
	for i := 0; i != len(*data); i++ {
		if (*data)[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageDropboxDeleteChunk(t *testing.T) {
	b, _ := (&StorageDropbox{}).NewBackend(*dropboxURL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(dropboxURL)

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Int())
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	shasumdata := sha256.Sum256(rnddata)
	shasum := hex.EncodeToString(shasumdata[:])

	_, err := b.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}

	_, err = b.LoadChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}

	err = b.DeleteChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}

	_, err = b.LoadChunk(shasum, part, totalParts)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
