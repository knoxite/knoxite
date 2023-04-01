/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2021,      Nicolas Martin <penguwin@penguwin.eu>
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

func findFiles(rootPath string, excludes []string) <-chan ArchiveResult {
	c := make(chan ArchiveResult)
	go func() {
		defer close(c)
		err := filepath.Walk(rootPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				c <- ArchiveResult{&Archive{Path: path}, err}
				if fi != nil && fi.IsDir() {
					return filepath.SkipDir
				}

				// if errors.Is(err, fs.ErrPermission) ||
				// 	errors.Is(err, fs.ErrNotExist) ||
				// 	errors.Is(err, os.ErrDeadlineExceeded) {
				// fmt.Fprintf(os.Stderr, "Could not find %s\n", path)
				return nil
				// }
			}
			if fi == nil {
				// fmt.Fprintf(os.Stderr, "Could not read %s\n", path)
				c <- ArchiveResult{&Archive{Path: path}, fmt.Errorf("%s: could not read", path)}
				return nil
			}

			match := false
			for _, exclude := range excludes {
				// fmt.Println("Matching", path, filepath.Base(path), exclude)
				match, err = filepath.Match(strings.ToLower(exclude), strings.ToLower(path))
				if err != nil {
					return fmt.Errorf("Invalid exclude filter '%s': %v", exclude, err)
				}
				if !match {
					match, _ = filepath.Match(strings.ToLower(exclude), strings.ToLower(filepath.Base(path)))
				}

				if match {
					// fmt.Printf("\n\nSkipping %s as it matches filter: %s\n\n", path, exclude)
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
				c <- ArchiveResult{
					Archive: &Archive{Path: path},
					Error:   &os.PathError{Op: "stat", Path: path, Err: errors.New("error reading metadata")},
				}
				return nil
			}
			archive := Archive{
				Path:    path,
				Mode:    fi.Mode(),
				ModTime: fi.ModTime().Unix(),
				UID:     statT.uid(),
				GID:     statT.gid(),
				// AbsPath: path,
				// FileInfo: fi,
			}
			if isSymLink(fi) {
				symlink, err := os.Readlink(path)
				if err != nil {
					c <- ArchiveResult{&archive, fmt.Errorf("error resolving symlink for: %v - %v", path, err)}
					return nil
				}

				archive.Type = SymLink
				archive.PointsTo = symlink
			} else if fi.IsDir() {
				archive.Type = Directory
			} else if isRegularFile(fi) {
				archive.Type = File
				archive.Size = uint64(fi.Size())
			} else {
				return nil
			}

			c <- ArchiveResult{Archive: &archive, Error: nil}
			return nil
		})

		if err != nil {
			c <- ArchiveResult{Archive: &Archive{Path: rootPath}, Error: err}
		}
	}()
	return c
}

func isSpecialPath(path string) bool {
	return path == "." || path == ".."
}

func isSymLink(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&os.ModeSymlink != 0
}

func isRegularFile(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&(os.ModeType|os.ModeCharDevice|os.ModeSymlink) == 0
}
