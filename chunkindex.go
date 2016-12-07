/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"encoding/json"
	"fmt"
)

// A ChunkIndexItem links a chunk with one or many snapshots
type ChunkIndexItem struct {
	ShaSum      string   `json:"sha256"`
	DataParts   uint     `json:"data_parts"`
	ParityParts uint     `json:"parity_parts"`
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
		decb, derr := Decrypt(b, repository.Password)
		err = derr
		if err == nil {
			err = json.Unmarshal(decb, &index)
		}
	} else {
		fmt.Println("Chunk-Index is empty, re-indexing now...")
		err = index.reindex(repository)
		if err == nil {
			err = index.Save(repository)
		}
	}
	return index, err
}

// Save writes a chunk-index
func (index *ChunkIndex) Save(repository *Repository) error {
	b, err := json.Marshal(*index)
	if err != nil {
		return err
	}

	encb, err := Encrypt(b, repository.Password)
	if err == nil {
		err = repository.Backend.SaveChunkIndex(encb)
	}
	return err
}

// Pack deletes unreferenced chunks and removes them from the index
func (index *ChunkIndex) Pack(repository *Repository) error {
	chunks := []*ChunkIndexItem{}

	for _, chunk := range index.Chunks {
		//	fmt.Printf("Chunk %s referenced in Snapshots %+v\n", chunk.ShaSum, chunk.Snapshots)
		if len(chunk.Snapshots) == 0 {
			fmt.Printf("Chunk %s is no longer referenced by any snapshot. Deleting!\n", chunk.ShaSum)

			for i := uint(0); i < chunk.DataParts+chunk.ParityParts; i++ {
				err := repository.Backend.DeleteChunk(chunk.ShaSum, i, chunk.DataParts)
				if err != nil {
					return err
				}
			}
		} else {
			chunks = append(chunks, chunk)
		}
	}

	index.Chunks = chunks
	return nil
}

func (index *ChunkIndex) reindex(repository *Repository) error {
	for _, vol := range repository.Volumes {
		for _, snapshotID := range vol.Snapshots {
			snapshot, err := vol.LoadSnapshot(snapshotID, repository)
			if err != nil {
				return err
			}

			for _, item := range snapshot.Items {
				index.AddItem(&item, snapshot.ID)
			}
		}
	}

	return nil
}

// AddItem updates chunk-index with the new chunks
func (index *ChunkIndex) AddItem(id *ItemData, snapshot string) {
	for _, chunk := range id.Chunks {
		found := false
		for _, c := range index.Chunks {
			if chunk.ShaSum == c.ShaSum {
				found = true
				c.Snapshots = append(c.Snapshots, snapshot)
			}
		}

		if !found {
			index.Chunks = append(index.Chunks, &ChunkIndexItem{chunk.ShaSum, chunk.DataParts, chunk.ParityParts, []string{snapshot}})
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
