/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"fmt"
)

// Progress contains stats and current path
type Progress struct {
	Path        string
	Size        uint64
	StorageSize uint64
}

func newProgress(item *ItemData) Progress {
	return Progress{
		Path:        item.Path,
		Size:        item.Size,
		StorageSize: item.StorageSize,
	}
}

// Stat contains a bunch of stats counters
type Stat struct {
	Files       uint64 `json:"files"`
	Dirs        uint64 `json:"dirs"`
	SymLinks    uint64 `json:"symlinks"`
	Size        uint64 `json:"size"`
	StorageSize uint64 `json:"stored_size"`
	Errors      uint64 `json:"errors"`
}

// Add accumulates other into s
func (s *Stat) Add(other Stat) {
	s.Files += other.Files
	s.Dirs += other.Dirs
	s.SymLinks += other.SymLinks
	s.Size += other.Size
	s.StorageSize += other.StorageSize
	s.Errors += other.Errors
}

// AddItem accumulates item into s
func (s *Stat) AddItem(i *ItemData) {
	s.Size += i.Size
	s.StorageSize += i.StorageSize

	switch i.Type {
	case SymLink:
		s.SymLinks++
	case File:
		s.Files++
	case Directory:
		s.Dirs++
	}
}

// SizeToString prettifies sizes
func SizeToString(size uint64) (str string) {
	b := float64(size)

	switch {
	case size > 1<<60:
		str = fmt.Sprintf("%.3f EiB", b/(1<<60))
	case size > 1<<50:
		str = fmt.Sprintf("%.3f PiB", b/(1<<50))
	case size > 1<<40:
		str = fmt.Sprintf("%.3f TiB", b/(1<<40))
	case size > 1<<30:
		str = fmt.Sprintf("%.3f GiB", b/(1<<30))
	case size > 1<<20:
		str = fmt.Sprintf("%.3f MiB", b/(1<<20))
	case size > 1<<10:
		str = fmt.Sprintf("%.3f KiB", b/(1<<10))
	default:
		str = fmt.Sprintf("%dB", size)
	}

	return
}

// String returns human-readable stats
func (s Stat) String() string {
	return fmt.Sprintf("%d files, %d dirs, %d symlinks, %d errors, %v Original Size, %v Storage Size",
		s.Files, s.Dirs, s.SymLinks, s.Errors, SizeToString(s.Size), SizeToString(s.StorageSize))
}
