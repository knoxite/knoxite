/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

// A Snapshot is a compilation of one or many archives
// MUST BE encrypted
type Snapshot struct {
	mut sync.Mutex

	ID          string              `json:"id"`
	Date        time.Time           `json:"date"`
	Description string              `json:"description"`
	Stats       Stats               `json:"stats"`
	Archives    map[string]*Archive `json:"items"`
}

// NewSnapshot creates a new snapshot
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

func IsInAnyPath(p string, paths []string) bool {
	for _, path := range paths {
		rel, _ := filepath.Rel(path, p)
		if !filepath.HasPrefix(rel, "../") {
			return true
		}
	}
	return false
}

func (snapshot *Snapshot) gatherTargetInformation(cwd string, paths []string, excludes []string, excludeExternalSymlinks bool, out chan ArchiveResult) {
	var wg sync.WaitGroup
	for _, path := range paths {
		c := findFiles(path, excludes)

		for result := range c {
			if result.Error == nil {
				rel, err := filepath.Rel(cwd, result.Archive.Path)
				if err == nil && !strings.HasPrefix(rel, "../") {
					result.Archive.Path = rel

				}
				if isSpecialPath(result.Archive.Path) {
					continue
				}
				if result.Archive.Type == SymLink {
					var relativeToCwd string
					if filepath.IsAbs(result.Archive.PointsTo) {
						relativeToCwd, _ = filepath.Rel(cwd, result.Archive.PointsTo)
					} else {
						var relativeToFile string
						if filepath.IsAbs(result.Archive.Path) {
							relativeToFile = filepath.Join(filepath.Dir(result.Archive.Path), result.Archive.PointsTo)
						} else {
							relativeToFile = filepath.Join(filepath.Join(cwd, filepath.Dir(result.Archive.Path)), result.Archive.PointsTo)
						}
						relativeToCwd, _ = filepath.Rel(cwd, relativeToFile)
					}

					relativePaths := []string{}
					for _, toRelativePath := range paths {
						relativePath, _ := filepath.Rel(cwd, toRelativePath)
						relativePaths = append(relativePaths, relativePath)
					}

					if err == nil && !IsInAnyPath(relativeToCwd, relativePaths) {
						if excludeExternalSymlinks {

							snapshot.mut.Lock()
							snapshot.Stats.Excluded++
							snapshot.mut.Unlock()

							continue
						}
					}
				}

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
func (snapshot *Snapshot) Add(cwd string, paths []string, excludes []string, repository Repository, chunkIndex *ChunkIndex, compress, encrypt uint16, dataParts, parityParts uint, excludeExternalSymlinks bool) chan Progress {
	progress := make(chan Progress)
	fwd := make(chan ArchiveResult)

	go snapshot.gatherTargetInformation(cwd, paths, excludes, excludeExternalSymlinks, fwd)

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
			snapshot.mut.Lock()
			p.TotalStatistics = snapshot.Stats
			snapshot.mut.Unlock()
			progress <- p

			if archive.Type == File {
				dataParts = uint(math.Max(1, float64(dataParts)))
				chunkchan, err := chunkFile(archive.Path, compress, encrypt, repository.password, int(dataParts), int(parityParts))
				if err != nil {
					if os.IsNotExist(err) {
						// if this file has already been deleted before we could backup it, we can gracefully ignore it and continue
						continue
					}
					p = newProgressError(err)
					progress <- p
					break
				}
				archive.Encrypted = encrypt
				archive.Compressed = compress

				for cd := range chunkchan {
					if cd.Error != nil {
						p = newProgressError(err)
						progress <- p
						close(progress)
						return
					}
					chunk := cd.Chunk
					// fmt.Printf("\tSplit %s (#%d, %d bytes), compression: %s, encryption: %s, hash: %s\n", id.Path, cd.Num, cd.Size, CompressionText(cd.Compressed), EncryptionText(cd.Encrypted), cd.Hash)

					// store this chunk
					n, err := repository.backend.StoreChunk(chunk)
					if err != nil {
						p = newProgressError(err)
						progress <- p
						close(progress)
						return
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
	snapshot := Snapshot{
		Archives: make(map[string]*Archive),
	}
	b, err := repository.backend.LoadSnapshot(id)
	if err != nil {
		return &snapshot, err
	}
	pipe, err := NewDecodingPipeline(CompressionLZMA, EncryptionAES, repository.password)
	if err != nil {
		return &snapshot, err
	}
	err = pipe.Decode(b, &snapshot)
	return &snapshot, err
}

// Save writes a snapshot's metadata
func (snapshot *Snapshot) Save(repository *Repository) error {
	pipe, err := NewEncodingPipeline(CompressionLZMA, EncryptionAES, repository.password)
	if err != nil {
		return err
	}
	b, err := pipe.Encode(snapshot)
	if err != nil {
		return err
	}
	return repository.backend.SaveSnapshot(snapshot.ID, b)
}

// AddArchive adds an archive to a snapshot
func (snapshot *Snapshot) AddArchive(archive *Archive) {
	snapshot.Archives[archive.Path] = archive
}
