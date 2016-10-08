// +build darwin dragonfly freebsd linux netbsd openbsd solaris

/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import "syscall"

// AvailableSpace returns the free space on this backend
func (backend *StorageLocal) AvailableSpace() (uint64, error) {
	//FIXME: make this cross-platform compatible
	var stat syscall.Statfs_t
	syscall.Statfs(backend.path, &stat)

	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize), nil
}
