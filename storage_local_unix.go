//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "syscall"

// AvailableSpace returns the free space on this backend.
func (backend *StorageLocal) AvailableSpace() (uint64, error) {
	//nolint:godox
	//FIXME: make this cross-platform compatible
	var stat syscall.Statfs_t
	err := syscall.Statfs(backend.Path, &stat)
	if err != nil {
		return 0, err
	}

	// Available blocks * size per block = available space in bytes
	// we convert both types to a uint64 as their type varies on different OS
	return stat.Bavail * uint64(stat.Bsize), nil
}
