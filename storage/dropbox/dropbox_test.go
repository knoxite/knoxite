// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package dropbox

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
	"testing"

	knoxite "github.com/knoxite/knoxite/lib"
	"github.com/knoxite/knoxite/storage"
)

var (
	backendTests []*storage.BackendTest
)

func TestMain(m *testing.M) {
	flag.Parse()

	// create a random bucket name every time to avoid collisions
	rnd := make([]byte, 8)
	rand.Read(rnd)

	dropboxurl := os.Getenv("KNOXITE_DROPBOX_URL")
	if len(dropboxurl) > 0 {
		backendTests = append(backendTests, &storage.BackendTest{
			URL:         dropboxurl + hex.EncodeToString(rnd),
			Protocols:   []string{"dropbox"},
			Description: "Dropbox Storage",
			TearDown: func(tb *storage.BackendTest) {
				db := tb.Backend.(*StorageDropbox)
				err := db.DeleteFile(db.Path)
				if err != nil {
					panic(err)
				}
			},
		})
	}

	if len(backendTests) == 0 {
		panic("no backends enabled")
	}
	for _, tt := range backendTests {
		var err error
		tt.Backend, err = knoxite.BackendFromURL(tt.URL)
		if err != nil {
			panic(err)
		}
	}

	r := m.Run()
	for _, tt := range backendTests {
		tt.TearDown(tt)
	}

	os.Exit(r)
}

func TestStorageNewBackend(t *testing.T) {
	storage.StorageNewBackendTest(t, backendTests)
}

func TestStorageLocation(t *testing.T) {
	storage.StorageLocationTest(t, backendTests)
}

func TestStorageProtocols(t *testing.T) {
	storage.StorageProtocolsTest(t, backendTests)
}

func TestStorageDescription(t *testing.T) {
	storage.StorageDescriptionTest(t, backendTests)
}

func TestStorageInitRepository(t *testing.T) {
	storage.StorageInitRepositoryTest(t, backendTests)
}

func TestStorageSaveRepository(t *testing.T) {
	storage.StorageSaveRepositoryTest(t, backendTests)
}

func TestStorageSaveSnapshot(t *testing.T) {
	storage.StorageSaveSnapshotTest(t, backendTests)
}

func TestStorageStoreChunk(t *testing.T) {
	storage.StorageStoreChunkTest(t, backendTests)
}

func TestStorageDeleteChunk(t *testing.T) {
	storage.StorageDeleteChunkTest(t, backendTests)
}
