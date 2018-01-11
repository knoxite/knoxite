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
	"strings"
)

func findFiles(rootPath string, excludes []string) chan ArchiveResult {
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

			match := false
			for _, exclude := range excludes {
				//fmt.Println("Matching", path, filepath.Base(path), exclude)
				match, err = filepath.Match(strings.ToLower(exclude), strings.ToLower(path))
				if err != nil {
					fmt.Println("Invalid exclude filter:", exclude)
					return err
				}
				if !match {
					match, _ = filepath.Match(strings.ToLower(exclude), strings.ToLower(filepath.Base(path)))
				}

				if match {
					fmt.Printf("Skipping %s as it matches filter: %s\n", path, exclude)
					break
				}
			}
			if match {
				if fi.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

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
