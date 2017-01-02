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

// A Snapshot is a compilation of one or many archives
// MUST BE encrypted
type Snapshot struct {
	sync.Mutex

	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Stats       Stats     `json:"stats"`
	Archives    []Archive `json:"items"`
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

func (snapshot *Snapshot) gatherTargetInformation(cwd string, paths []string, out chan ArchiveResult) {
	var wg sync.WaitGroup
	for _, path := range paths {
		c := findFiles(path)

		for result := range c {
			if result.Error == nil {
				rel, err := filepath.Rel(cwd, result.Archive.Path)
				if err == nil && !strings.HasPrefix(rel, "../") {
					result.Archive.Path = rel
				}
				if isSpecialPath(result.Archive.Path) {
					continue
				}

				snapshot.Lock()
				snapshot.Stats.Size += result.Archive.Size
				snapshot.Unlock()
			}

			wg.Add(1)
			go func(r ArchiveResult) {
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
	fwd := make(chan ArchiveResult)

	go snapshot.gatherTargetInformation(cwd, paths, fwd)

	go func() {
		for result := range fwd {
			if result.Error != nil {
				p := newProgressError(result.Error)
				progress <- p
				break
			}

			archive := result.Archive
			rel, err := filepath.Rel(cwd, archive.Path)
			if err == nil && !strings.HasPrefix(rel, "../") {
				archive.Path = rel
			}
			if isSpecialPath(archive.Path) {
				continue
			}

			p := newProgress(archive)
			snapshot.Lock()
			p.TotalStatistics = snapshot.Stats
			snapshot.Unlock()
			progress <- p

			if isRegularFile(archive.FileInfo) {
				dataParts = uint(math.Max(1, float64(dataParts)))
				chunkchan, err := chunkFile(archive.AbsPath, compress, encrypt, repository.Password, int(dataParts), int(parityParts))
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

					archive.Chunks = append(archive.Chunks, cd)
					archive.StorageSize += n

					p.CurrentItemStats.StorageSize = archive.StorageSize
					p.CurrentItemStats.Transferred += uint64(cd.OriginalSize)
					snapshot.Stats.Transferred += uint64(cd.OriginalSize)
					snapshot.Stats.StorageSize += n

					snapshot.Lock()
					p.TotalStatistics = snapshot.Stats
					snapshot.Unlock()
					progress <- p
				}
			}

			snapshot.AddArchive(archive)
			chunkIndex.AddArchive(archive, snapshot.ID)
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
	s.Archives = snapshot.Archives

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
		b, err = Uncompress(b)
		if err != nil {
			return &snapshot, err
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
		b, err = Compress(b)
		if err != nil {
			return err
		}
	}

	b, err = Encrypt(b, repository.Password)
	if err == nil {
		err = repository.Backend.SaveSnapshot(snapshot.ID, b)
	}
	return err
}

// AddArchive adds an archive to a snapshot
func (snapshot *Snapshot) AddArchive(archive *Archive) {
	archives := []Archive{}

	found := false
	for _, v := range snapshot.Archives {
		if v.Path == archive.Path {
			found = true

			archives = append(archives, *archive)
		} else {
			archives = append(archives, v)
		}
	}

	if !found {
		archives = append(archives, *archive)
	}

	snapshot.Archives = archives
}
