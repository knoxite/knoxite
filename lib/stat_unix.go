// +build darwin dragonfly freebsd linux netbsd openbsd solaris

/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "syscall"

type statUnix syscall.Stat_t

func (s statUnix) dev() uint64   { return uint64(s.Dev) }
func (s statUnix) ino() uint64   { return s.Ino }
func (s statUnix) nlink() uint64 { return uint64(s.Nlink) }
func (s statUnix) uid() uint32   { return s.Uid }
func (s statUnix) gid() uint32   { return s.Gid }
func (s statUnix) rdev() uint64  { return uint64(s.Rdev) }
func (s statUnix) size() int64   { return s.Size }

func toStatT(i interface{}) (statT, bool) {
	if i == nil {
		return nil, false
	}
	s, ok := i.(*syscall.Stat_t)
	if ok && s != nil {
		return statUnix(*s), true
	}
	return nil, false
}
