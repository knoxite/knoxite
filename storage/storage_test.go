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
	"net/url"
	"os"
	"reflect"
	"testing"

	knoxite "github.com/knoxite/knoxite/lib"

	_ "github.com/knoxite/knoxite/storage/amazons3"
	"github.com/knoxite/knoxite/storage/backblaze"
	"github.com/knoxite/knoxite/storage/dropbox"
	"github.com/knoxite/knoxite/storage/ftp"
)

type testBackend struct {
	url         string
	protocols   []string
	description string
	tearDown    func(tb *testBackend)
}

var (
	testBackends []*testBackend
)

func TestMain(m *testing.M) {
	flag.Parse()

	// create a random bucket name every time to avoid collisions
	rnd := make([]byte, 8)
	rand.Read(rnd)

	backblazeurl := os.Getenv("KNOXITE_BACKBLAZE_URL")
	if len(backblazeurl) > 0 {
		testBackends = append(testBackends, &testBackend{
			url:         backblazeurl + hex.EncodeToString(rnd),
			protocols:   []string{"backblaze"},
			description: "Backblaze Storage",
			tearDown: func(tb *testBackend) {
				b, err := knoxite.BackendFromURL(tb.url)
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
		testBackends = append(testBackends, &testBackend{
			url:         dropboxurl + hex.EncodeToString(rnd),
			protocols:   []string{"dropbox"},
			description: "Dropbox Storage",
			tearDown: func(tb *testBackend) {
				b, err := knoxite.BackendFromURL(tb.url)
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

	amazons3url := os.Getenv("KNOXITE_AMAZONS3_URL")
	if len(amazons3url) > 0 {
		testBackends = append(testBackends, &testBackend{
			url:         amazons3url + hex.EncodeToString(rnd),
			protocols:   []string{"s3", "s3s"},
			description: "Amazon S3 Storage",
			tearDown: func(tb *testBackend) {
				// create a random bucket name every time to avoid collisions
				rnd = make([]byte, 8)
				rand.Read(rnd)

				tb.url = amazons3url + hex.EncodeToString(rnd)
			},
		})
	}

	ftpurl := os.Getenv("KNOXITE_FTP_URL")
	if len(ftpurl) > 0 {
		testBackends = append(testBackends, &testBackend{
			url:         ftpurl,
			protocols:   []string{"ftp"},
			description: "FTP Storage",
			tearDown: func(tb *testBackend) {
				b, err := knoxite.BackendFromURL(tb.url)
				if err != nil {
					panic(err)
				}

				u, err := url.Parse(tb.url)
				if err != nil {
					panic(err)
				}

				db := b.(*ftp.StorageFTP)
				err = db.DeletePath(u.Path)
				if err != nil {
					panic(err)
				}

				err = db.CreatePath(u.Path)
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
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err := b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)
	}
}

func TestStorageSaveRepository(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnd := make([]byte, 256)
		rand.Read(rnd)

		err = b.SaveRepository(rnd)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadRepository()
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if !reflect.DeepEqual(data, rnd) {
			t.Errorf("%s: Data missmatch", tt.description)
		}
	}
}

func TestStorageLoadRepository(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnd := make([]byte, 256)
		rand.Read(rnd)

		err = b.SaveRepository(rnd)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadRepository()
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if !reflect.DeepEqual(data, rnd) {
			t.Errorf("%s: Data missmatch", tt.description)
		}
	}
}

func TestStorageSaveSnapshot(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		rndid := make([]byte, 8)
		rand.Read(rndid)
		id := hex.EncodeToString(rndid)

		err = b.SaveSnapshot(id, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadSnapshot(id)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data missmatch", tt.description)
		}
	}
}

func TestStorageLoadSnapshot(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		rndid := make([]byte, 8)
		rand.Read(rndid)
		id := hex.EncodeToString(rndid)

		err = b.SaveSnapshot(id, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadSnapshot(id)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data missmatch", tt.description)
		}
	}
}

func TestStorageStoreChunk(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Int())
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		shasumdata := sha256.Sum256(rnddata)
		shasum := hex.EncodeToString(shasumdata[:])

		size, err := b.StoreChunk(shasum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if size != uint64(len(rnddata)) {
			t.Errorf("%s: Data length missmatch", tt.description)
		}

		// Test to store the same chunk twice. Size should be 0
		size, err = b.StoreChunk(shasum, part, totalParts, rnddata)
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
		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data missmatch", tt.description)
		}
	}
}

func TestStorageLoadChunk(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Int())
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		shasumdata := sha256.Sum256(rnddata)
		shasum := hex.EncodeToString(shasumdata[:])

		_, err = b.StoreChunk(shasum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := b.LoadChunk(shasum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data missmatch", tt.description)
		}
	}
}

func TestStorageDeleteChunk(t *testing.T) {
	for _, tt := range testBackends {
		b, err := knoxite.BackendFromURL(tt.url)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
			continue
		}

		if err = b.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		defer tt.tearDown(tt)

		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Int())
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		shasumdata := sha256.Sum256(rnddata)
		shasum := hex.EncodeToString(shasumdata[:])

		_, err = b.StoreChunk(shasum, part, totalParts, rnddata)
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
