// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package storage

import (
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"
	"reflect"
	"testing"

	knoxite "github.com/knoxite/knoxite/lib"
)

type BackendTest struct {
	URL         string
	Protocols   []string
	Description string
	Backend     knoxite.Backend
	TearDown    func(tb *BackendTest)
}

func StorageNewBackendTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		_, err := knoxite.BackendFromURL(tt.URL)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}
	}
}

func StorageLocationTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		if tt.Backend.Location() != tt.URL {
			t.Errorf("%s: Expected %v, got %v", tt.Description, tt.URL, tt.Backend.Location())
		}
	}
}

func StorageProtocolsTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		if len(tt.Backend.Protocols()) != len(tt.Protocols) {
			t.Errorf("%s: Invalid amount of protocols", tt.Description)
		}

		for i := 0; i < len(tt.Protocols); i++ {
			if tt.Backend.Protocols()[i] != tt.Protocols[i] {
				t.Errorf("%s: Invalid protocols", tt.Description)
			}
		}
	}
}

func StorageDescriptionTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		if tt.Backend.Description() != tt.Description {
			t.Errorf("%s: Invalid Description", tt.Description)
		}
	}
}

func StorageInitRepositoryTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		if err := tt.Backend.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}
	}
}

func StorageSaveRepositoryTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		rnd := make([]byte, 256)
		rand.Read(rnd)

		err := tt.Backend.SaveRepository(rnd)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}

		data, err := tt.Backend.LoadRepository()
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}

		if !reflect.DeepEqual(data, rnd) {
			t.Errorf("%s: Data mismatch %d %d", tt.Description, len(data), len(rnd))
		}
	}
}

func StorageSaveSnapshotTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		rndid := make([]byte, 8)
		rand.Read(rndid)
		id := hex.EncodeToString(rndid)

		err := tt.Backend.SaveSnapshot(id, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}

		data, err := tt.Backend.LoadSnapshot(id)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}

		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data mismatch", tt.Description)
		}
	}
}

func StorageStoreChunkTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Intn(256))
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		hashsum := knoxite.Hash(rnddata, knoxite.HashHighway256)
		size, err := tt.Backend.StoreChunk(hashsum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}
		if size != uint64(len(rnddata)) {
			t.Errorf("%s: Data length mismatch: %d != %d", tt.Description, size, len(rnddata))
		}

		// Test to store the same chunk twice. Size should be 0
		size, err = tt.Backend.StoreChunk(hashsum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}
		if size != 0 {
			t.Errorf("%s: Already exisiting chunks should not be overwritten", tt.Description)
		}

		data, err := tt.Backend.LoadChunk(hashsum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}
		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data mismatch", tt.Description)
		}
	}
}

func StorageDeleteChunkTest(t *testing.T, backendTests []*BackendTest) {
	for _, tt := range backendTests {
		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Intn(256))
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		hashsum := knoxite.Hash(rnddata, knoxite.HashHighway256)
		_, err := tt.Backend.StoreChunk(hashsum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}

		err = tt.Backend.DeleteChunk(hashsum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.Description, err)
		}

		_, err = tt.Backend.LoadChunk(hashsum, part, totalParts)
		if err == nil {
			t.Errorf("%s: Expected error, got nil", tt.Description)
		}
	}
}
