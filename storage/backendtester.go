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

	"github.com/knoxite/knoxite"
)

type BackendTest struct {
	URL         string
	Protocols   []string
	Description string

	Backend  knoxite.Backend
	TearDown func(tb *BackendTest)
}

func RandomSuffix() string {
	rnd := make([]byte, 8)
	rand.Read(rnd)

	return hex.EncodeToString(rnd)
}

func RunBackendTester(b *BackendTest, m *testing.M) {
	flag.Parse()

	var err error
	b.Backend, err = knoxite.BackendFromURL(b.URL)
	if err != nil {
		panic(err)
	}

	r := m.Run()
	b.TearDown(b)

	os.Exit(r)
}

func (b *BackendTest) NewBackendTest(t *testing.T) {
	_, err := knoxite.BackendFromURL(b.URL)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
}

func (b *BackendTest) LocationTest(t *testing.T) {
	// since urls get parsed in BackendFromURL and some protocols use url-encoded parameters,
	// we need to parse the url in LocationTest, too
	expectedLocation, err := url.Parse(b.URL)
	if err != nil {
		panic(err)
	}
	if b.Backend.Location() != expectedLocation.String() {
		t.Errorf("%s: Expected %v, got %v", b.Description, expectedLocation.String(), b.Backend.Location())
	}
}

func (b *BackendTest) ProtocolsTest(t *testing.T) {
	if len(b.Backend.Protocols()) != len(b.Protocols) {
		t.Errorf("%s: Invalid amount of protocols", b.Description)
	}

	for i := 0; i < len(b.Protocols); i++ {
		if b.Backend.Protocols()[i] != b.Protocols[i] {
			t.Errorf("%s: Invalid protocols", b.Description)
		}
	}
}

func (b *BackendTest) DescriptionTest(t *testing.T) {
	if b.Backend.Description() != b.Description {
		t.Errorf("%s: Invalid Description", b.Description)
	}
}

func (b *BackendTest) InitRepositoryTest(t *testing.T) {
	if err := b.Backend.InitRepository(); err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
}

func (b *BackendTest) SaveRepositoryTest(t *testing.T) {
	rnd := make([]byte, 256)
	rand.Read(rnd)

	err := b.Backend.SaveRepository(rnd)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	data, err := b.Backend.LoadRepository()
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	if !reflect.DeepEqual(data, rnd) {
		t.Errorf("%s: Data mismatch %d %d", b.Description, len(data), len(rnd))
	}
}

func (b *BackendTest) AvailableSpaceTest(t *testing.T) {
	space, err := b.Backend.AvailableSpace()
	if err != nil && err != knoxite.ErrAvailableSpaceUnknown && err != knoxite.ErrAvailableSpaceUnlimited {
		t.Errorf("%s: expected available space information, got %s", b.Description, err)
	}
	if err == nil && space <= 0 {
		t.Errorf("%s: expected available space information, got %d", b.Description, space)
	}
}

func (b *BackendTest) SaveSnapshotTest(t *testing.T) {
	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	rndid := make([]byte, 8)
	rand.Read(rndid)
	id := hex.EncodeToString(rndid)

	err := b.Backend.SaveSnapshot(id, rnddata)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	data, err := b.Backend.LoadSnapshot(id)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	if !reflect.DeepEqual(data, rnddata) {
		t.Errorf("%s: Data mismatch", b.Description)
	}
}

func (b *BackendTest) StoreChunkTest(t *testing.T) {
	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Intn(256))
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	hashsum := knoxite.Hash(rnddata, knoxite.HashHighway256)
	size, err := b.Backend.StoreChunk(hashsum, part, totalParts, rnddata)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
	if size != uint64(len(rnddata)) {
		t.Errorf("%s: Data length mismatch: %d != %d", b.Description, size, len(rnddata))
	}

	// Test to store the same chunk twice. Size should be 0
	size, err = b.Backend.StoreChunk(hashsum, part, totalParts, rnddata)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
	if size != 0 {
		t.Errorf("%s: Already exisiting chunks should not be overwritten", b.Description)
	}

	data, err := b.Backend.LoadChunk(hashsum, part, totalParts)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
	if !reflect.DeepEqual(data, rnddata) {
		t.Errorf("%s: Data mismatch", b.Description)
	}
}

func (b *BackendTest) DeleteChunkTest(t *testing.T) {
	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Intn(256))
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	hashsum := knoxite.Hash(rnddata, knoxite.HashHighway256)
	_, err := b.Backend.StoreChunk(hashsum, part, totalParts, rnddata)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	err = b.Backend.DeleteChunk(hashsum, part, totalParts)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	_, err = b.Backend.LoadChunk(hashsum, part, totalParts)
	if err == nil {
		t.Errorf("%s: Expected error, got nil", b.Description)
	}
}

func (b *BackendTest) LockingTest(t *testing.T) {
	lock := []byte("lock")

	// test locking repository
	l, err := b.Backend.LockRepository(lock)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
	if len(l) > 0 {
		t.Errorf("%s: Expected empty lock, got %d bytes", b.Description, len(l))
	}

	// try locking while already locked
	lnew, err := b.Backend.LockRepository(lock)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
	// should return the old lock
	if !reflect.DeepEqual(lock, lnew) {
		t.Errorf("%s: Data mismatch", b.Description)
	}

	// unlock
	err = b.Backend.UnlockRepository()
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}

	// lock again
	l, err = b.Backend.LockRepository(lock)
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
	if len(l) > 0 {
		t.Errorf("%s: Expected empty lock, got %d bytes", b.Description, len(l))
	}

	// clean up lock
	err = b.Backend.UnlockRepository()
	if err != nil {
		t.Errorf("%s: %s", b.Description, err)
	}
}
