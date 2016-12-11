package knoxite

import "time"

// Progress contains stats and current path
type Progress struct {
	Path        string
	Size        uint64
	StorageSize uint64
	Transferred uint64
	Timer       time.Time
	Statistics  Stats
	Error       error
}

func newProgress(item *ItemData) Progress {
	return Progress{
		Path:        item.Path,
		Size:        item.Size,
		StorageSize: item.StorageSize,
		Timer:       time.Now(),
		Statistics:  Stats{},
		Error:       nil,
	}
}

func newProgressError(err error) Progress {
	return Progress{
		Error: err,
	}
}

// TransferSpeed returns the average transfer speed in bytes per second
func (p Progress) TransferSpeed() uint64 {
	return uint64(float64(p.Transferred) / time.Since(p.Timer).Seconds())
}
