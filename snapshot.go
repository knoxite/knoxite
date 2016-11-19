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
func (snapshot *Snapshot) Add(cwd string, paths []string, repository Repository, compress, encrypt bool, dataParts, parityParts uint) chan Progress {
	progress := make(chan Progress)
	fwd := make(chan ItemResult, 256) // TODO: reconsider buffer size
	m := new(sync.Mutex)
	var totalSize uint64 // total data size: uncompressed, unencrypted

	go func() {
		for _, path := range paths {
			c := findFiles(path)

			for result := range c {
				if result.Error == nil {
					rel, err := filepath.Rel(cwd, result.Item.Path)
					if err == nil && !strings.HasPrefix(rel, "../") {
						result.Item.Path = rel
					}
					if isSpecialPath(result.Item.Path) {
						continue
					}
					m.Lock()
					totalSize += result.Item.Size
					m.Unlock()
				}
				fwd <- result
			}
		}
		close(fwd)
	}()

	go func() {
		var totalTransferredSize uint64 // total transferred size: uncompressed, unencrypted
		var totalStorageSize uint64     // total storage size: compressed, encrypted
		for result := range fwd {
			if result.Error != nil {
				p := newProgressError(result.Error)
				progress <- p
				break
			}

			item := result.Item
			rel, err := filepath.Rel(cwd, item.Path)
			if err == nil && !strings.HasPrefix(rel, "../") {
				item.Path = rel
			}
			if isSpecialPath(item.Path) {
				continue
			}

			p := newProgress(item)
			m.Lock()
			p.Statistics.Size = totalSize
			p.Statistics.StorageSize = totalStorageSize
			m.Unlock()
			progress <- p

			if isRegularFile(item.FileInfo) {
				var currentTransferredSize uint64 // current file's transferred size: uncompressed, unencrypted
				dataParts = uint(math.Max(1, float64(dataParts)))
				chunkchan, err := chunkFile(item.AbsPath, compress, encrypt, repository.Password, int(dataParts), int(parityParts))
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

					item.Chunks = append(item.Chunks, cd)
					item.StorageSize += n
					totalStorageSize += n
					currentTransferredSize += uint64(cd.OriginalSize)
					totalTransferredSize += uint64(cd.OriginalSize)

					p := newProgress(item)
					m.Lock()
					p.Transferred = currentTransferredSize
					p.Statistics.Size = totalSize
					p.Statistics.StorageSize = totalStorageSize
					p.Statistics.Transferred = totalTransferredSize
					m.Unlock()
					progress <- p
				}
			}

			snapshot.AddItem(item)
		}
		close(progress)
	}()

	return progress
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
