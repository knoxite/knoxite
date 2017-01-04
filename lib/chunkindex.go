/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// A ChunkIndexItem links a chunk with one or many snapshots
type ChunkIndexItem struct {
	ShaSum      string   `json:"sha256"`
	DataParts   uint     `json:"data_parts"`
	ParityParts uint     `json:"parity_parts"`
	Size        int      `json:"size"`
	Snapshots   []string `json:"snapshots"`
}

// A ChunkIndex links chunks with snapshots
// MUST BE encrypted
type ChunkIndex struct {
	Chunks []*ChunkIndexItem `json:"chunks"`
}

// OpenChunkIndex opens an existing chunkindex
func OpenChunkIndex(repository *Repository) (ChunkIndex, error) {
	index := ChunkIndex{}
	b, err := repository.Backend.LoadChunkIndex()
	if err == nil {
		r, errd := Decrypt(bytes.NewReader(b), repository.Password)
		if errd != nil {
			return index, errd
		}
		defer r.Close()

		if repository.Version == 1 {
			r, err = Uncompress(r)
			if err != nil {
				return index, err
			}
			defer r.Close()
		}

		b, ierr := ioutil.ReadAll(r)
		if ierr != nil {
			return index, ierr
		}
		err = json.Unmarshal(b, &index)
	} else {
		if !repository.IsEmpty() {
			fmt.Println("Chunk-Index is empty, re-indexing all snapshots...")
			err = index.reindex(repository)
			if err == nil {
				if len(index.Chunks) > 0 {
					fmt.Println("Successfully re-indexed snapshots.")
				}
			}
		}

		err = index.Save(repository)
	}
	return index, err
}

// Save writes a chunk-index
func (index *ChunkIndex) Save(repository *Repository) error {
	b, err := json.Marshal(*index)
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
		err = repository.Backend.SaveChunkIndex(b)
	}
	return err
}

// Pack deletes unreferenced chunks and removes them from the index
func (index *ChunkIndex) Pack(repository *Repository) (freedSize uint64, err error) {
	chunks := []*ChunkIndexItem{}

	for _, chunk := range index.Chunks {
		//	fmt.Printf("Chunk %s referenced in Snapshots %+v\n", chunk.ShaSum, chunk.Snapshots)
		if len(chunk.Snapshots) == 0 {
			fmt.Printf("Chunk %s is no longer referenced by any snapshot. Deleting!\n", chunk.ShaSum)

			for i := uint(0); i < chunk.DataParts+chunk.ParityParts; i++ {
				err = repository.Backend.DeleteChunk(chunk.ShaSum, i, chunk.DataParts)
				if err != nil {
					return
				}
				freedSize += uint64(chunk.Size)
			}
		} else {
			chunks = append(chunks, chunk)
		}
	}

	index.Chunks = chunks
	return
}

func (index *ChunkIndex) reindex(repository *Repository) error {
	for _, vol := range repository.Volumes {
		for _, snapshotID := range vol.Snapshots {
			snapshot, err := vol.LoadSnapshot(snapshotID, repository)
			if err != nil {
				return err
			}

			for _, archive := range snapshot.Archives {
				index.AddArchive(&archive, snapshot.ID)
			}
		}
	}

	return nil
}

// AddArchive updates chunk-index with the new chunks
func (index *ChunkIndex) AddArchive(archive *Archive, snapshot string) {
	for _, chunk := range archive.Chunks {
		found := false
		for _, c := range index.Chunks {
			if chunk.ShaSum == c.ShaSum {
				found = true
				c.Snapshots = append(c.Snapshots, snapshot)
			}
		}

		if !found {
			chunkItem := ChunkIndexItem{
				ShaSum:      chunk.ShaSum,
				DataParts:   chunk.DataParts,
				ParityParts: chunk.ParityParts,
				Size:        chunk.Size,
				Snapshots:   []string{snapshot},
			}
			index.Chunks = append(index.Chunks, &chunkItem)
		}
	}
}

// RemoveSnapshot removes all references to snapshot from the chunk-index
func (index *ChunkIndex) RemoveSnapshot(snapshot string) {
	for _, chunk := range index.Chunks {
		snapshots := []string{}
		for _, s := range chunk.Snapshots {
			if s != snapshot {
				snapshots = append(snapshots, s)
			}
		}

		chunk.Snapshots = snapshots
	}
}
