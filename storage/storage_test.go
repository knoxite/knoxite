/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	mrand "math/rand"
	"os"
	"testing"

	knoxite "github.com/knoxite/knoxite/lib"

	_ "github.com/knoxite/knoxite/storage/amazons3"

	backblaze "github.com/knoxite/knoxite/storage/backblaze"
	dropbox "github.com/knoxite/knoxite/storage/dropbox"

	_ "github.com/knoxite/knoxite/storage/ftp"
	_ "github.com/knoxite/knoxite/storage/http"
)

type testBackend struct {
	url         string
	protocols   []string
	description string
	tearDown    func(url string)
}

var (
	testBackends []testBackend
)

func TestMain(m *testing.M) {
	flag.Parse()

	backblazeurl := os.Getenv("KNOXITE_BACKBLAZE_URL")
	if len(backblazeurl) > 0 {
		testBackends = append(testBackends, testBackend{
			url:         backblazeurl,
			protocols:   []string{"backblaze"},
			description: "Backblaze Storage",
			tearDown: func(url string) {
				b, err := knoxite.BackendFromURL(url)
				if err != nil {
					panic(err)
				}

				db := b.(*backblaze.StorageBackblaze)
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

	dropboxurl := os.Getenv("KNOXITE_DROPBOX_URL")
	if len(dropboxurl) > 0 {
		testBackends = append(testBackends, testBackend{
			url:         dropboxurl,
			protocols:   []string{"dropbox"},
			description: "Dropbox Storage",
			tearDown: func(url string) {
				b, err := knoxite.BackendFromURL(url)
				if err != nil {
					panic(err)
				}

				db := b.(*dropbox.StorageDropbox)
				err = db.DeleteFile(db.Path)
				if err != nil {
					panic(err)
				}
			},
		})
	}

	os.Exit(m.Run())
}

func TestStorageNewBackend(t *testing.T) {
	for _, tt := range testBackends {
		_, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
	}
}

func TestStorageLocation(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)

		if b.Location() != tt.url {
			t.Errorf("%s: Expected %v, got %v", tt.description, tt.url, b.Location())
		}
	}
}

func TestStorageProtocols(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if len(b.Protocols()) != len(tt.protocols) {
			t.Errorf("%s: Invalid amount of protocols", tt.description)
		}

		for i := 0; i < len(tt.protocols); i++ {
			if b.Protocols()[i] != tt.protocols[i] {
				t.Errorf("%s: Invalid protocols", tt.description)
			}
		}
	}
}

func TestStorageDescription(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)

		if b.Description() != tt.description {
			t.Errorf("%s: Invalid Description", tt.description)
		}
	}
}

func TestStorageInitRepository(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)

		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)
	}
}

func TestStorageSaveRepository(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnd := make([]byte, 256)
		rand.Read(rnd)

		err := b.SaveRepository(rnd)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadRepository()
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if len(data) != len(rnd) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}

		for i := 0; i != len(data); i++ {
			if data[i] != rnd[i] {
				t.Errorf("%s: Data missmatch", tt.description)
			}
		}
	}
}

func TestStorageLoadRepository(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnd := make([]byte, 256)
		rand.Read(rnd)

		err := b.SaveRepository(rnd)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadRepository()
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if len(data) != len(rnd) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}

		for i := 0; i != len(data); i++ {
			if data[i] != rnd[i] {
				t.Errorf("%s: Data missmatch", tt.description)
			}
		}
	}
}

func TestStorageSaveSnapshot(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		rndid := make([]byte, 8)
		rand.Read(rndid)
		id := hex.EncodeToString(rndid)

		err := b.SaveSnapshot(id, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadSnapshot(id)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if len(data) != len(rnddata) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}

		for i := 0; i != len(data); i++ {
			if data[i] != rnddata[i] {
				t.Errorf("%s: Data missmatch", tt.description)
			}
		}
	}
}

func TestStorageLoadSnapshot(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		rndid := make([]byte, 8)
		rand.Read(rndid)
		id := hex.EncodeToString(rndid)

		err := b.SaveSnapshot(id, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadSnapshot(id)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if len(data) != len(rnddata) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}

		for i := 0; i != len(data); i++ {
			if data[i] != rnddata[i] {
				t.Errorf("%s: Data missmatch", tt.description)
			}
		}
	}
}

func TestStorageStoreChunk(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Int())
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		shasumdata := sha256.Sum256(rnddata)
		shasum := hex.EncodeToString(shasumdata[:])

		size, err := b.StoreChunk(shasum, part, totalParts, &rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if size != uint64(len(rnddata)) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}

		// Test to store the same chunk twice. Size should be 0
		size, err = b.StoreChunk(shasum, part, totalParts, &rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if size != 0 {
			t.Errorf("%s: Already exisiting chunks should not be overwritten", tt.description)
		}

		data, err := b.LoadChunk(shasum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if len(*data) != len(rnddata) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}
		for i := 0; i != len(*data); i++ {
			if (*data)[i] != rnddata[i] {
				t.Errorf("%s: Data missmatch", tt.description)
			}
		}
	}
}

func TestStorageLoadChunk(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Int())
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		shasumdata := sha256.Sum256(rnddata)
		shasum := hex.EncodeToString(shasumdata[:])

		_, err := b.StoreChunk(shasum, part, totalParts, &rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadChunk(shasum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if len(*data) != len(rnddata) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}
		for i := 0; i != len(*data); i++ {
			if (*data)[i] != rnddata[i] {
				t.Errorf("%s: Data missmatch", tt.description)
			}
		}
	}
}

func TestStorageDeleteChunk(t *testing.T) {
	for _, tt := range testBackends {
		b, _ := knoxite.BackendFromURL(tt.url)
		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt.url)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Int())
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		shasumdata := sha256.Sum256(rnddata)
		shasum := hex.EncodeToString(shasumdata[:])

		_, err := b.StoreChunk(shasum, part, totalParts, &rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		_, err = b.LoadChunk(shasum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		err = b.DeleteChunk(shasum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		_, err = b.LoadChunk(shasum, part, totalParts)
		if err == nil {
			t.Errorf("%s: Expected error, got nil", tt.description)
		}
	}
}
