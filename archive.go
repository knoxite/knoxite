/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io"
	"os"
)

// Which type
const (
	File      = iota // A File
	Directory        // A Directory
	SymLink          // A SymLink
)

// Archive contains all metadata belonging to a file/directory.
type Archive struct {
	Path        string      `json:"path"`               // Where in filesystem does this belong to
	PointsTo    string      `json:"pointsto,omitempty"` // If this is a SymLink, where does it point to
	Mode        os.FileMode `json:"mode"`               // file mode bits
	ModTime     int64       `json:"modtime"`            // modification time
	Size        uint64      `json:"size"`               // size
	StorageSize uint64      `json:"storagesize"`        // size in storage
	UID         uint32      `json:"uid"`                // owner
	GID         uint32      `json:"gid"`                // group
	Chunks      []Chunk     `json:"chunks,omitempty"`   // data chunks
	Encrypted   uint16      `json:"encrypted"`          // encryption type
	Compressed  uint16      `json:"compressed"`         // compression type
	Type        uint8       `json:"type"`               // Is this a File, Directory or SymLink
}

// ArchiveResult wraps Archive and an error.
// Either Archive or Error is nil.
type ArchiveResult struct {
	Archive *Archive
	Error   error
}

// IndexOfChunk returns the slice-index for a specific chunk number.
func (arc *Archive) IndexOfChunk(chunkNum uint) (int, error) {
	for i, chunk := range arc.Chunks {
		if chunk.Num == chunkNum {
			return i, nil
		}
	}

	return 0, &ChunkError{chunkNum}
}

// ChunkForOffset returns the chunk containing data beginning at offset.
// Returns chunk-number, offset inside this chunk, error.
func (arc *Archive) ChunkForOffset(offset int) (uint, int, error) {
	size := 0
	for i := 0; i < len(arc.Chunks); i++ {
		idx, err := arc.IndexOfChunk(uint(i))
		if err != nil {
			return 0, 0, &SeekError{offset}
		}

		chunk := arc.Chunks[idx]
		if size+chunk.OriginalSize > offset {
			internalOffset := offset - size
			return chunk.Num, internalOffset, nil
		}

		size += chunk.OriginalSize
	}

	return 0, 0, io.EOF
}
