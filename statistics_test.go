/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "testing"

func TestStatisticsAdd(t *testing.T) {
	s := []Stats{}

	for i := uint64(1); i <= 3; i++ {
		v := Stats{
			Files:       i,
			Dirs:        i,
			SymLinks:    i,
			Size:        i,
			StorageSize: i,
			Transferred: i,
			Errors:      i,
		}

		s = append(s, v)
	}

	s[0].Add(s[1])
	if s[0] != s[2] {
		t.Errorf("Expected %v, got %v", s[2], s[0])
	}
}

func TestStatisticsString(t *testing.T) {
	tests := []struct {
		size        uint64
		storageSize uint64
		result      string
	}{
		{1, 1, "1 files, 2 dirs, 3 symlinks, 6 errors, 1B Original Size, 1B Storage Size"},
		{1024, 1024, "1 files, 2 dirs, 3 symlinks, 6 errors, 1.00 KiB Original Size, 1.00 KiB Storage Size"},
		{1024 * (1 << 10), 1024 * (1 << 10), "1 files, 2 dirs, 3 symlinks, 6 errors, 1.00 MiB Original Size, 1.00 MiB Storage Size"},
		{1024 * (1 << 20), 1024 * (1 << 20), "1 files, 2 dirs, 3 symlinks, 6 errors, 1.00 GiB Original Size, 1.00 GiB Storage Size"},
		{1024 * (1 << 30), 1024 * (1 << 30), "1 files, 2 dirs, 3 symlinks, 6 errors, 1.00 TiB Original Size, 1.00 TiB Storage Size"},
		{1024 * (1 << 40), 1024 * (1 << 40), "1 files, 2 dirs, 3 symlinks, 6 errors, 1.00 PiB Original Size, 1.00 PiB Storage Size"},
		{1024 * (1 << 50), 1024 * (1 << 50), "1 files, 2 dirs, 3 symlinks, 6 errors, 1.00 EiB Original Size, 1.00 EiB Storage Size"},
	}
	for _, tt := range tests {
		s := Stats{
			Files:       1,
			Dirs:        2,
			SymLinks:    3,
			Size:        tt.size,
			StorageSize: tt.storageSize,
			Errors:      6,
		}

		v := s.String()
		if v != tt.result {
			t.Errorf("Expected %s, got %s", tt.result, v)
		}
	}
}
