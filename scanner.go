/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
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

// ItemData contains all metadata belonging to a file/directory
// MUST BE encrypted
type ItemData struct {
	Path        string      `json:"path"`               // Where in filesystem does this belong to
	Type        uint        `json:"type"`               // Is this a File, Directory or SymLink
	PointsTo    string      `json:"pointsto,omitempty"` // If this is a SymLink, where does it point to
	Mode        os.FileMode `json:"mode"`               // file mode bits
	ModTime     time.Time   `json:"modtime"`            // modification time
	Size        uint64      `json:"size"`               // size
	StorageSize uint64      `json:"storagesize"`        // size in storage
	UID         uint32      `json:"uid"`                // owner
	GID         uint32      `json:"gid"`                // group
	Chunks      []Chunk     `json:"chunks,omitempty"`
	AbsPath     string      `json:"-"`
	FileInfo    os.FileInfo `json:"-"`
}

func findFiles(rootPath string) chan ItemData {
	c := make(chan ItemData)
	go func() {
		err := filepath.Walk(rootPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find %s\n", path)
				return err
			}
			if fi == nil {
				fmt.Fprintf(os.Stderr, "Could not read %s\n", path)
				return fmt.Errorf("error for %v: FileInfo is nil", path)
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
			id := ItemData{
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

				id.Type = SymLink
				id.PointsTo = symlink
			} else if fi.IsDir() {
				id.Type = Directory
			} else {
				id.Type = File
				if isRegularFile(fi) {
					id.Size = uint64(fi.Size())
				}
			}

			c <- id
			return nil
		})

		//FIXME: handle errors gracefully
		if err != nil {
			panic(err)
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
