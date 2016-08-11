/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"encoding/json"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

// A Snapshot is compiled by one or many archives
// MUST BE encrypted
type Snapshot struct {
	ID          string     `json:"id"`
	Date        time.Time  `json:"date"`
	Description string     `json:"description"`
	Stats       Stats      `json:"stats"`
	Items       []ItemData `json:"items"`
}

// NewSnapshot creates a new snapshot
func NewSnapshot(description string) (Snapshot, error) {
	snapshot := Snapshot{
		Date:        time.Now(),
		Description: description,
	}

	u, err := uuid.NewV4()
	if err != nil {
		return snapshot, err
	}
	snapshot.ID = u.String()[:8]

	return snapshot, nil
}

// Add adds a path to a Snapshot
func (snapshot *Snapshot) Add(cwd, path string, repository Repository, compress, encrypt bool, dataParts, parityParts uint) (chan Progress, error) {
	progress := make(chan Progress)
	c := findFiles(path)
	fwd := make(chan ItemData, 256) // TODO: reconsider buffer size
	m := new(sync.Mutex)
	var totalSize uint64 = 0

	go func() {
		for id := range c {
			rel, err := filepath.Rel(cwd, id.Path)
			if err == nil && !strings.HasPrefix(rel, "../") {
				id.Path = rel
			}
			if isSpecialPath(id.Path) {
				continue
			}
			m.Lock()
			totalSize += id.Size
			m.Unlock()
			fwd <- id
		}
		close(fwd)
	}()

	go func() {
		for id := range fwd {
			rel, err := filepath.Rel(cwd, id.Path)
			if err == nil && !strings.HasPrefix(rel, "../") {
				id.Path = rel
			}
			if isSpecialPath(id.Path) {
				continue
			}

			progress <- newProgress(&id)

			if isRegularFile(id.FileInfo) {
				dataParts = uint(math.Max(1, float64(dataParts)))
				chunkchan, err := chunkFile(id.AbsPath, compress, encrypt, repository.Password, int(dataParts), int(parityParts))
				if err != nil {
					panic(err)
				}
				for cd := range chunkchan {
					// fmt.Printf("\tSplit %s (#%d, %d bytes), compression: %s, encryption: %s, sha256: %s\n", id.Path, cd.Num, cd.Size, CompressionText(cd.Compressed), EncryptionText(cd.Encrypted), cd.ShaSum)

					// store this chunk
					n, err := repository.Backend.StoreChunk(cd)
					if err != nil {
						panic(err)
					}

					// release the memory, we don't need the data anymore
					cd.Data = &[][]byte{}

					id.Chunks = append(id.Chunks, cd)
					id.StorageSize += n

					p := newProgress(&id)
					m.Lock()
					p.Statistics.StorageSize = totalSize
					m.Unlock()
					progress <- p
				}
			}

			snapshot.AddItem(&id)
		}
		close(progress)
	}()
	return progress, nil
}

// Clone clones a snapshot
func (snapshot *Snapshot) Clone() (*Snapshot, error) {
	s, err := NewSnapshot(snapshot.Description)
	if err != nil {
		return &s, err
	}

	s.Stats = snapshot.Stats
	s.Items = snapshot.Items

	return &s, nil
}

// OpenSnapshot opens an existing snapshot
func openSnapshot(id string, repository *Repository) (Snapshot, error) {
	snapshot := Snapshot{}
	b, err := repository.Backend.LoadSnapshot(id)

	decb, err := Decrypt(b, repository.Password)
	if err == nil {
		err = json.Unmarshal(decb, &snapshot)
	}
	return snapshot, err
}

// Save writes a snapshot's metadata
func (snapshot *Snapshot) Save(repository *Repository) error {
	//	b, err := json.MarshalIndent(*r, "", "    ")
	b, err := json.Marshal(*snapshot)
	if err != nil {
		return err
	}
	//	fmt.Printf("Repository created: %s\n", string(b))

	encb, err := Encrypt(b, repository.Password)
	if err == nil {
		err = repository.Backend.SaveSnapshot(snapshot.ID, encb)
	}
	return err
}

// AddItem adds an item to a snapshot
func (snapshot *Snapshot) AddItem(id *ItemData) {
	items := []ItemData{}
	stats := Stats{}

	found := false
	for _, i := range snapshot.Items {
		if i.Path == id.Path {
			found = true

			items = append(items, *id)
			stats.AddItem(id)
		} else {
			items = append(items, i)
			stats.AddItem(&i)
		}
	}

	if !found {
		items = append(items, *id)
		stats.AddItem(id)
	}

	snapshot.Items = items
	snapshot.Stats = stats
}
