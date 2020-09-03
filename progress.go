/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "time"

// Progress contains stats and current path.
type Progress struct {
	Path             string
	Timer            time.Time
	CurrentItemStats Stats
	TotalStatistics  Stats
	Verbosity        string
	Error            error
}

func newProgress(archive *Archive) Progress {
	return Progress{
		Path:  archive.Path,
		Timer: time.Now(),
		CurrentItemStats: Stats{
			Size:        archive.Size,
			StorageSize: archive.StorageSize,
		},
		TotalStatistics: Stats{
			Size: archive.Size,
		},
		Verbosity: "Verbosity flag has been set",
		Error:     nil,
	}
}

func newProgressError(err error) Progress {
	return Progress{
		Error: err,
	}
}

// TransferSpeed returns the average transfer speed in bytes per second.
func (p Progress) TransferSpeed() uint64 {
	return uint64(float64(p.CurrentItemStats.Transferred) / time.Since(p.Timer).Seconds())
}
