//go:build windows
// +build windows

/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

// AvailableSpace returns the free space on this backend
func (backend *StorageLocal) AvailableSpace() (uint64, error) {
	//nolint:godox
	//FIXME: make this cross-platform compatible
	return 0, nil
}
