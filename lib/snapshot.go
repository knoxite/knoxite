/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

// A Snapshot is a compilation of one or many archives
// MUST BE encrypted
type Snapshot struct {
	sync.Mutex

	ID          string     `json:"id"`
	Date        time.Time  `json:"date"`
	Description string     `json:"description"`
	Stats       Stats      `json:"stats"`
	Items       []ItemData `json:"items"`
}

// NewSnapshot creates a new snapshot
func NewSnapshot(description string) (*Snapshot, error) {
	snapshot := Snapshot{
		Date:        time.Now(),
		Description: description,
	}

	u, err := uuid.NewV4()
	if err != nil {
		return &snapshot, err
	}
	snapshot.ID = u.String()[:8]

	return &snapshot, nil
}

func (snapshot *Snapshot) gatherTargetInformation(cwd string, paths []string, out chan ItemResult) {
	var wg sync.WaitGroup
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

				snapshot.Lock()
				snapshot.Stats.Size += result.Item.Size
				snapshot.Unlock()
			}

			wg.Add(1)
			go func(r ItemResult) {
				out <- r
				wg.Done()
			}(result)
		}
	}

	wg.Wait()
	close(out)
}

// Add adds a path to a Snapshot
func (snapshot *Snapshot) Add(cwd string, paths []string, repository Repository, chunkIndex *ChunkIndex, compress, encrypt bool, dataParts, parityParts uint) chan Progress {
	progress := make(chan Progress)
	fwd := make(chan ItemResult)

	go snapshot.gatherTargetInformation(cwd, paths, fwd)

	go func() {
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
			snapshot.Lock()
			p.TotalStatistics = snapshot.Stats
			snapshot.Unlock()
			progress <- p

			if isRegularFile(item.FileInfo) {
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

					p.CurrentItemStats.StorageSize = item.StorageSize
					p.CurrentItemStats.Transferred += uint64(cd.OriginalSize)
					snapshot.Stats.Transferred += uint64(cd.OriginalSize)
					snapshot.Stats.StorageSize += n

					snapshot.Lock()
					p.TotalStatistics = snapshot.Stats
					snapshot.Unlock()
					progress <- p
				}
			}

			snapshot.AddItem(item)
			chunkIndex.AddItem(item, snapshot.ID)
		}
		close(progress)
	}()

	return progress
}

// Clone clones a snapshot
func (snapshot *Snapshot) Clone() (*Snapshot, error) {
	s, err := NewSnapshot(snapshot.Description)
	if err != nil {
		return s, err
	}

	s.Stats = snapshot.Stats
	s.Items = snapshot.Items

	return s, nil
}

// openSnapshot opens an existing snapshot
func openSnapshot(id string, repository *Repository) (*Snapshot, error) {
	snapshot := Snapshot{}
	b, err := repository.Backend.LoadSnapshot(id)
	if err != nil {
		return &snapshot, err
	}

	b, err = Decrypt(b, repository.Password)
	if err != nil {
		return &snapshot, err
	}

	if repository.Version == 1 {
		zr, zerr := gzip.NewReader(bytes.NewReader(b))
		if zerr != nil {
			return &snapshot, zerr
		}
		defer zr.Close()
		b, zerr = ioutil.ReadAll(zr)
		if zerr != nil {
			return &snapshot, zerr
		}
	}

	err = json.Unmarshal(b, &snapshot)
	return &snapshot, err
}

// Save writes a snapshot's metadata
func (snapshot *Snapshot) Save(repository *Repository) error {
	b, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	if repository.Version == 1 {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		w.Write(b)
		w.Close()
		b = buf.Bytes()
	}

	b, err = Encrypt(b, repository.Password)
	if err == nil {
		err = repository.Backend.SaveSnapshot(snapshot.ID, b)
	}
	return err
}

// AddItem adds an item to a snapshot
func (snapshot *Snapshot) AddItem(id *ItemData) {
	items := []ItemData{}

	found := false
	for _, i := range snapshot.Items {
		if i.Path == id.Path {
			found = true

			items = append(items, *id)
		} else {
			items = append(items, i)
		}
	}

	if !found {
		items = append(items, *id)
	}

	snapshot.Items = items
}
