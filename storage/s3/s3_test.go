// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package s3

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

	amazons3url := os.Getenv("KNOXITE_AMAZONS3_URL")
	if len(amazons3url) > 0 {
		backendTests = append(backendTests, &storage.BackendTest{
			URL:         amazons3url + hex.EncodeToString(rnd),
			Protocols:   []string{"s3", "s3s"},
			Description: "Amazon S3 Storage",
			TearDown: func(tb *storage.BackendTest) {
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
