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

// Stats contains a bunch of Stats counters
type Stats struct {
	Files       uint64 `json:"files"`
	Dirs        uint64 `json:"dirs"`
	SymLinks    uint64 `json:"symlinks"`
	Size        uint64 `json:"size"`
	StorageSize uint64 `json:"stored_size"`
	Transferred uint64 `json:"transferred"`
	Errors      uint64 `json:"errors"`
}

// Add accumulates other into s
func (s *Stats) Add(other Stats) {
	s.Files += other.Files
	s.Dirs += other.Dirs
	s.SymLinks += other.SymLinks
	s.Size += other.Size
	s.StorageSize += other.StorageSize
	s.Transferred += other.Transferred
	s.Errors += other.Errors
}

// SizeToString prettifies sizes
func SizeToString(size uint64) (str string) {
	b := float64(size)

	switch {
	case size >= 1<<60:
		str = fmt.Sprintf("%.2f EiB", b/(1<<60))
	case size >= 1<<50:
		str = fmt.Sprintf("%.2f PiB", b/(1<<50))
	case size >= 1<<40:
		str = fmt.Sprintf("%.2f TiB", b/(1<<40))
	case size >= 1<<30:
		str = fmt.Sprintf("%.2f GiB", b/(1<<30))
	case size >= 1<<20:
		str = fmt.Sprintf("%.2f MiB", b/(1<<20))
	case size >= 1<<10:
		str = fmt.Sprintf("%.2f KiB", b/(1<<10))
	default:
		str = fmt.Sprintf("%dB", size)
	}

	return
}

// String returns human-readable Stats
func (s Stats) String() string {
	return fmt.Sprintf("%d files, %d dirs, %d symlinks, %d errors, %v Original Size, %v Storage Size",
		s.Files, s.Dirs, s.SymLinks, s.Errors, SizeToString(s.Size), SizeToString(s.StorageSize))
}
