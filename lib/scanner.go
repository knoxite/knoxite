/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Which type
const (
	File      = iota // A File
	Directory        // A Directory
	SymLink          // A SymLink
)

// Archive contains all metadata belonging to a file/directory
// MUST BE encrypted
type Archive struct {
	Path        string      `json:"path"`               // Where in filesystem does this belong to
	Type        uint        `json:"type"`               // Is this a File, Directory or SymLink
	PointsTo    string      `json:"pointsto,omitempty"` // If this is a SymLink, where does it point to
	Mode        os.FileMode `json:"mode"`               // file mode bits
	ModTime     time.Time   `json:"modtime"`            // modification time
	Size        uint64      `json:"size"`               // size
	StorageSize uint64      `json:"storagesize"`        // size in storage
	UID         uint32      `json:"uid"`                // owner
	GID         uint32      `json:"gid"`                // group
	Chunks      []Chunk     `json:"chunks,omitempty"`   // data chunks
	AbsPath     string      `json:"-"`                  // Absolute path
	FileInfo    os.FileInfo `json:"-"`                  // FileInfo struct
}

// ArchiveResult wraps Archive and an error
// Either Archive or Error is nil
type ArchiveResult struct {
	Archive *Archive
	Error   error
}

func findFiles(rootPath string) chan ArchiveResult {
	c := make(chan ArchiveResult)
	go func() {
		err := filepath.Walk(rootPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				// fmt.Fprintf(os.Stderr, "Could not find %s\n", path)
				return err
			}
			if fi == nil {
				// fmt.Fprintf(os.Stderr, "Could not read %s\n", path)
				return fmt.Errorf("%s: could not read", path)
			}

			/* if !isExcluded(str, fi) {
				if fi.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}*/

			statT, ok := toStatT(fi.Sys())
			if !ok {
				return &os.PathError{Op: "stat", Path: path, Err: errors.New("error reading metadata")}
			}
			archive := Archive{
				Path:     path,
				AbsPath:  path,
				Mode:     fi.Mode(),
				ModTime:  fi.ModTime(),
				UID:      statT.uid(),
				GID:      statT.gid(),
				FileInfo: fi,
			}
			if isSymLink(fi) {
				symlink, lerr := os.Readlink(path)
				if lerr != nil {
					//FIXME: we should probably even (re)store invalid symlinks
					fmt.Fprintf(os.Stderr, "error resolving symlink for: %v - %v\n", path, lerr)
					return nil
				}

				archive.Type = SymLink
				archive.PointsTo = symlink
			} else if fi.IsDir() {
				archive.Type = Directory
			} else {
				archive.Type = File
				if isRegularFile(fi) {
					archive.Size = uint64(fi.Size())
				}
			}

			c <- ArchiveResult{Archive: &archive, Error: nil}
			return nil
		})

		if err != nil {
			c <- ArchiveResult{Archive: nil, Error: err}
		}
		close(c)
	}()
	return c
}

func isSpecialPath(path string) bool {
	return path == "."
}

func isSymLink(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&os.ModeSymlink != 0
}

func isRegularFile(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&(os.ModeType|os.ModeCharDevice|os.ModeSymlink) == 0
}
