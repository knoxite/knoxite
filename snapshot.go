/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2021,      Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

// A Snapshot is a compilation of one or many archives.
type Snapshot struct {
	mut sync.Mutex

	ID          string              `json:"id"`
	Date        time.Time           `json:"date"`
	Description string              `json:"description"`
	Stats       Stats               `json:"stats"`
	Archives    map[string]*Archive `json:"items"`
}

// StoreOptions holds all the storage settings for a snapshot operation.
type StoreOptions struct {
	CWD         string
	Paths       []string
	Excludes    []string
	Compress    uint16
	Encrypt     uint16
	Pedantic    bool
	DataParts   uint
	ParityParts uint
	Verify      bool
}

// NewSnapshot creates a new snapshot.
func NewSnapshot(description string) (*Snapshot, error) {
	snapshot := Snapshot{
		Date:        time.Now(),
		Description: description,
		Archives:    make(map[string]*Archive),
	}

	u, err := uuid.NewV4()
	if err != nil {
		return &snapshot, err
	}
	snapshot.ID = u.String()[:8]

	return &snapshot, nil
}

func (snapshot *Snapshot) gatherTargetInformation(cwd string, paths []string, excludes []string) <-chan ArchiveResult {
	ch := make(chan ArchiveResult)
	var wg sync.WaitGroup

	results := func(rr []ArchiveResult) {
		go func() {
			for _, r := range rr {
				ch <- r
				wg.Done()
			}
		}()
	}

	go func() {
		var archives []ArchiveResult

		for _, path := range paths {
			ff := findFiles(path, excludes)

			for result := range ff {
				if result.Error == nil {
					rel, err := filepath.Rel(cwd, result.Archive.Path)
					if err == nil && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
						result.Archive.Path = rel
					}
					if isSpecialPath(result.Archive.Path) {
						continue
					}

					// update scan statistics
					snapshot.mut.Lock()
					snapshot.Stats.Size += result.Archive.Size
					switch result.Archive.Type {
					case Directory:
						snapshot.Stats.Dirs++
					case File:
						snapshot.Stats.Files++
					case SymLink:
						snapshot.Stats.SymLinks++
					}
					snapshot.mut.Unlock()
				}

				wg.Add(1)
				archives = append(archives, result)

				if len(archives) >= 128 {
					results(archives)
					archives = []ArchiveResult{}
				}
			}
		}

		results(archives)

		wg.Wait()
		close(ch)
	}()

	return ch
}

// Add adds a path to a Snapshot.
func (snapshot *Snapshot) Add(repository Repository, chunkIndex *ChunkIndex, opts StoreOptions) <-chan Progress {
	progress := make(chan Progress)

	ch := snapshot.gatherTargetInformation(opts.CWD, opts.Paths, opts.Excludes)

	go func() {
		defer close(progress)
		for result := range ch {
			if result.Error != nil {
				p := newProgressError(result.Error)
				p.Path = result.Archive.Path
				progress <- p
				if opts.Pedantic {
					break
				}
				continue
			}

			archive := result.Archive
			rel, err := filepath.Rel(opts.CWD, archive.Path)
			if err == nil && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				archive.Path = rel
			}
			if isSpecialPath(archive.Path) {
				continue
			}

			p := newProgress(archive)
			snapshot.mut.Lock()
			p.TotalStatistics = snapshot.Stats
			snapshot.mut.Unlock()
			progress <- p

			if archive.Type == File {
				opts.DataParts = uint(math.Max(1, float64(opts.DataParts)))
				chunkchan, err := chunkFile(archive.Path, repository.Key, opts)
				if err != nil {
					if os.IsNotExist(err) {
						// if this file has already been deleted before we could backup it, we can gracefully ignore it and continue
						continue
					}
					p = newProgressError(err)
					p.Path = archive.Path
					progress <- p
					if opts.Pedantic {
						break
					}
					continue
				}
				archive.Encrypted = opts.Encrypt
				archive.Compressed = opts.Compress

				for cd := range chunkchan {
					if cd.Error != nil {
						p = newProgressError(err)
						p.Path = archive.Path
						progress <- p
						if opts.Pedantic {
							return
						}
						continue
					}
					chunk := cd.Chunk
					// fmt.Printf("\tSplit %s (#%d, %d bytes), compression: %s, encryption: %s, hash: %s\n", id.Path, cd.Num, cd.Size, CompressionText(cd.Compressed), EncryptionText(cd.Encrypted), cd.Hash)

					// store this chunk
					n, err := repository.backend.StoreChunk(chunk)
					if err != nil {
						p = newProgressError(err)
						p.Path = archive.Path
						progress <- p
						if opts.Pedantic {
							return
						}
						continue
					}

					if opts.Verify {
						for i, data := range *chunk.Data {
							b, err := repository.backend.LoadChunk(chunk, uint(i))
							if err != nil {
								p = newProgressError(fmt.Errorf("Failed to re-load %s: %v", archive.Path, err))
								p.Path = archive.Path
								progress <- p
								if opts.Pedantic {
									return
								}
								continue
							}

							if !bytes.Equal(b, data) {
								p = newProgressError(fmt.Errorf("Stored and loaded chunks vary"))
								p.Path = archive.Path
								progress <- p
								if opts.Pedantic {
									return
								}
							}
						}
					}

					// release the memory, we don't need the data anymore
					chunk.Data = &[][]byte{}

					archive.Chunks = append(archive.Chunks, chunk)
					archive.StorageSize += n

					p.CurrentItemStats.StorageSize = archive.StorageSize
					p.CurrentItemStats.Transferred += uint64(chunk.OriginalSize)
					snapshot.Stats.Transferred += uint64(chunk.OriginalSize)
					snapshot.Stats.StorageSize += n

					snapshot.mut.Lock()
					p.TotalStatistics = snapshot.Stats
					snapshot.mut.Unlock()
					progress <- p
				}
			}

			snapshot.AddArchive(archive)
			chunkIndex.AddArchive(archive, snapshot.ID)
		}

	}()

	return progress
}

// Clone clones a snapshot.
func (snapshot *Snapshot) Clone() (*Snapshot, error) {
	s, err := NewSnapshot(snapshot.Description)
	if err != nil {
		return s, err
	}

	s.Stats = snapshot.Stats
	s.Archives = snapshot.Archives

	return s, nil
}

// openSnapshot opens an existing snapshot.
func openSnapshot(id string, repository *Repository) (*Snapshot, error) {
	snapshot := Snapshot{
		Archives: make(map[string]*Archive),
	}
	b, err := repository.backend.LoadSnapshot(id)
	if err != nil {
		return &snapshot, err
	}
	pipe, err := NewDecodingPipeline(CompressionLZMA, EncryptionAES, repository.Key)
	if err != nil {
		return &snapshot, err
	}
	err = pipe.Decode(b, &snapshot)
	return &snapshot, err
}

// Save writes a snapshot's metadata.
func (snapshot *Snapshot) Save(repository *Repository) error {
	pipe, err := NewEncodingPipeline(CompressionLZMA, EncryptionAES, repository.Key)
	if err != nil {
		return err
	}
	b, err := pipe.Encode(snapshot)
	if err != nil {
		return err
	}
	return repository.backend.SaveSnapshot(snapshot.ID, b)
}

// AddArchive adds an archive to a snapshot.
func (snapshot *Snapshot) AddArchive(archive *Archive) {
	snapshot.Archives[archive.Path] = archive
}
