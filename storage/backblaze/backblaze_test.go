// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package backblaze

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

	backblazeurl := os.Getenv("KNOXITE_BACKBLAZE_URL")
	if len(backblazeurl) > 0 {
		backendTests = append(backendTests, &storage.BackendTest{
			URL:         backblazeurl + hex.EncodeToString(rnd),
			Protocols:   []string{"backblaze"},
			Description: "Backblaze Storage",
			TearDown: func(tb *storage.BackendTest) {
				db := tb.Backend.(*StorageBackblaze)
				list, err := db.Bucket.ListFileNames("", 128)
				if err != nil {
					panic(err)
				}

				for _, v := range list.Files {
					ids, lerr := db.Bucket.ListFileVersions(v.Name, "", 128)
					if lerr != nil {
						panic(lerr)
					}
					for _, id := range ids.Files {
						db.Bucket.DeleteFileVersion(v.Name, id.ID)
					}
				}

				err = db.Bucket.Delete()
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
