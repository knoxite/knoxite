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
	testBackends []*storage.TestBackend
)

func TestMain(m *testing.M) {
	flag.Parse()

	// create a random bucket name every time to avoid collisions
	rnd := make([]byte, 8)
	rand.Read(rnd)

	backblazeurl := os.Getenv("KNOXITE_BACKBLAZE_URL")
	if len(backblazeurl) > 0 {
		testBackends = append(testBackends, &storage.TestBackend{
			URL:         backblazeurl + hex.EncodeToString(rnd),
			Protocols:   []string{"backblaze"},
			Description: "Backblaze Storage",
			TearDown: func(tb *storage.TestBackend) {
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

	if len(testBackends) == 0 {
		panic("no backends enabled")
	}
	for _, tt := range testBackends {
		var err error
		tt.Backend, err = knoxite.BackendFromURL(tt.URL)
		if err != nil {
			panic(err)
		}
	}

	r := m.Run()
	for _, tt := range testBackends {
		tt.TearDown(tt)
	}

	os.Exit(r)
}

func TestStorageNewBackend(t *testing.T) {
	storage.StorageNewBackendTest(t, testBackends)
}

func TestStorageLocation(t *testing.T) {
	storage.StorageLocationTest(t, testBackends)
}

func TestStorageProtocols(t *testing.T) {
	storage.StorageProtocolsTest(t, testBackends)
}

func TestStorageDescription(t *testing.T) {
	storage.StorageDescriptionTest(t, testBackends)
}

func TestStorageInitRepository(t *testing.T) {
	storage.StorageInitRepositoryTest(t, testBackends)
}

func TestStorageSaveRepository(t *testing.T) {
	storage.StorageSaveRepositoryTest(t, testBackends)
}

func TestStorageSaveSnapshot(t *testing.T) {
	storage.StorageSaveSnapshotTest(t, testBackends)
}

func TestStorageStoreChunk(t *testing.T) {
	storage.StorageStoreChunkTest(t, testBackends)
}

func TestStorageDeleteChunk(t *testing.T) {
	storage.StorageDeleteChunkTest(t, testBackends)
}
