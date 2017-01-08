/*
 * knoxite
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package backblaze

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
	b2URL *url.URL
)

func TestMain(m *testing.M) {
	flag.Parse()

	path := os.Getenv("KNOXITE_BACKBLAZE_URL")
	if len(path) == 0 {
		panic("KNOXITE_BACKBLAZE_URL is undefined")
	}
	// create a random bucket name every time to avoid collisions
	rnd := make([]byte, 8)
	rand.Read(rnd)
	path += hex.EncodeToString(rnd)

	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}

	b2URL = u
	os.Exit(m.Run())
}

func teardownRepo(u *url.URL) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	db := b.(*StorageBackblaze)
	list, err := db.bucket.ListFileNames("", 128)
	if err != nil {
		panic(err)
	}

	for _, v := range list.Files {
		ids, lerr := db.bucket.ListFileVersions(v.Name, "", 128)
		if lerr != nil {
			panic(lerr)
		}
		for _, id := range ids.Files {
			db.bucket.DeleteFileVersion(v.Name, id.ID)
		}
	}

	err = db.bucket.Delete()
	if err != nil {
		panic(err)
	}
}

func TestStorageBackblazeNewBackend(t *testing.T) {
	_, err := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err != nil {
		t.Error(err)
	}
}

func TestStorageBackblazeLocation(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if b.Location() != b2URL.String() {
		t.Errorf("Expected %v, got %v", b2URL.String(), b.Location())
	}
}

func TestStorageBackblazeProtocols(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if len(b.Protocols()) != 1 {
		t.Error("Invalid amount of protocols")
	}

	if b.Protocols()[0] != "backblaze" {
		t.Error("Invalid protocols")
	}
}

func TestStorageBackblazeDescription(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)

	if b.Description() != "Backblaze Storage" {
		t.Error("Invalid Description")
	}
}

func TestStorageBackblazeInitRepository(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)

	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)
}

func TestStorageBackblazeSaveRepository(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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

func TestStorageBackblazeLoadRepository(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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

func TestStorageBackblazeSaveSnapshot(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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

func TestStorageBackblazeLoadSnapshot(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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

func TestStorageBackblazeStoreChunk(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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

func TestStorageBackblazeLoadChunk(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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

func TestStorageBackblazeDeleteChunk(t *testing.T) {
	b, _ := (&StorageBackblaze{}).NewBackend(*b2URL)
	if err := b.InitRepository(); err != nil {
		t.Error(err)
	}
	defer teardownRepo(b2URL)

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
