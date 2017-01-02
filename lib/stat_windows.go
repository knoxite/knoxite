// +build windows

/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "syscall"

type statWin syscall.Win32FileAttributeData

func toStatT(i interface{}) (statT, bool) {
	if i == nil {
		return nil, false
	}
	s, ok := i.(*syscall.Win32FileAttributeData)
	if ok && s != nil {
		return statWin(*s), true
	}
	return nil, false
}

func (s statWin) dev() uint64   { return 0 }
func (s statWin) ino() uint64   { return 0 }
func (s statWin) nlink() uint64 { return 0 }
func (s statWin) uid() uint32   { return 0 }
func (s statWin) gid() uint32   { return 0 }
func (s statWin) rdev() uint64  { return 0 }

func (s statWin) size() int64 {
	return int64(s.FileSizeLow) | (int64(s.FileSizeHigh) << 32)
}
