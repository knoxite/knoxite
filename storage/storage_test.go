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
	"github.com/knoxite/knoxite/storage/sftp"
)

type testBackend struct {
	url         string
	protocols   []string
	description string
	backend     knoxite.Backend
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
				db := tb.backend.(*backblaze.StorageBackblaze)
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
				db := tb.backend.(*dropbox.StorageDropbox)
				err := db.DeleteFile(db.Path)
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
				u, err := url.Parse(tb.url)
				if err != nil {
					panic(err)
				}

				db := tb.backend.(*ftp.StorageFTP)
				err = db.DeletePath(u.Path)
				if err != nil {
					panic(err)
				}
			},
		})
	}

	sftpurl := os.Getenv("KNOXITE_SFTP_URL")
	if len(sftpurl) > 0 {
		testBackends = append(testBackends, &testBackend{
			url:         sftpurl,
			protocols:   []string{"sftp"},
			description: "SSH/SFTP Storage",
			tearDown: func(tb *testBackend) {
				u, err := url.Parse(tb.url)
				if err != nil {
					panic(err)
				}

				db := tb.backend.(*sftp.StorageSFTP)
				err = db.DeletePath(u.Path)
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
		tt.backend, err = knoxite.BackendFromURL(tt.url)
		if err != nil {
			panic(err)
		}
	}

	r := m.Run()

	for _, tt := range testBackends {
		tt.tearDown(tt)
	}
	os.Exit(r)
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
		if tt.backend.Location() != tt.url {
			t.Errorf("%s: Expected %v, got %v", tt.description, tt.url, tt.backend.Location())
		}
	}
}

func TestStorageProtocols(t *testing.T) {
	for _, tt := range testBackends {
		if len(tt.backend.Protocols()) != len(tt.protocols) {
			t.Errorf("%s: Invalid amount of protocols", tt.description)
		}

		for i := 0; i < len(tt.protocols); i++ {
			if tt.backend.Protocols()[i] != tt.protocols[i] {
				t.Errorf("%s: Invalid protocols", tt.description)
			}
		}
	}
}

func TestStorageDescription(t *testing.T) {
	for _, tt := range testBackends {
		if tt.backend.Description() != tt.description {
			t.Errorf("%s: Invalid Description", tt.description)
		}
	}
}

func TestStorageInitRepository(t *testing.T) {
	for _, tt := range testBackends {
		if err := tt.backend.InitRepository(); err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
	}
}

func TestStorageSaveRepository(t *testing.T) {
	for _, tt := range testBackends {
		rnd := make([]byte, 256)
		rand.Read(rnd)

		err := tt.backend.SaveRepository(rnd)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := tt.backend.LoadRepository()
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if !reflect.DeepEqual(data, rnd) {
			t.Errorf("%s: Data mismatch %d %d", tt.description, len(data), len(rnd))
		}
	}
}

func TestStorageSaveSnapshot(t *testing.T) {
	for _, tt := range testBackends {
		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		rndid := make([]byte, 8)
		rand.Read(rndid)
		id := hex.EncodeToString(rndid)

		err := tt.backend.SaveSnapshot(id, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		data, err := tt.backend.LoadSnapshot(id)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data mismatch", tt.description)
		}
	}
}

func TestStorageStoreChunk(t *testing.T) {
	for _, tt := range testBackends {
		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Intn(256))
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		hashsum := knoxite.Hash(rnddata, knoxite.HashHighway256)
		size, err := tt.backend.StoreChunk(hashsum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if size != uint64(len(rnddata)) {
			t.Errorf("%s: Data length mismatch: %d != %d", tt.description, size, len(rnddata))
		}

		// Test to store the same chunk twice. Size should be 0
		size, err = tt.backend.StoreChunk(hashsum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if size != 0 {
			t.Errorf("%s: Already exisiting chunks should not be overwritten", tt.description)
		}

		data, err := tt.backend.LoadChunk(hashsum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}
		if !reflect.DeepEqual(data, rnddata) {
			t.Errorf("%s: Data mismatch", tt.description)
		}
	}
}

func TestStorageDeleteChunk(t *testing.T) {
	for _, tt := range testBackends {
		rnddata := make([]byte, 256)
		rand.Read(rnddata)

		totalParts := uint(mrand.Intn(256))
		// get a random part number which is smaller than the totalParts number
		part := uint(mrand.Intn(int(totalParts)))

		hashsum := knoxite.Hash(rnddata, knoxite.HashHighway256)
		_, err := tt.backend.StoreChunk(hashsum, part, totalParts, rnddata)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		err = tt.backend.DeleteChunk(hashsum, part, totalParts)
		if err != nil {
			t.Errorf("%s: %s", tt.description, err)
		}

		_, err = tt.backend.LoadChunk(hashsum, part, totalParts)
		if err == nil {
			t.Errorf("%s: Expected error, got nil", tt.description)
		}
	}
}
