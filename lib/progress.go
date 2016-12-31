package knoxite

import "time"

// Progress contains stats and current path
type Progress struct {
	Path             string
	Timer            time.Time
	CurrentItemStats Stats
	TotalStatistics  Stats
	Error            error
}

func newProgress(item *ItemData) Progress {
	return Progress{
		Path:  item.Path,
		Timer: time.Now(),
		CurrentItemStats: Stats{
			Size:        item.Size,
			StorageSize: item.StorageSize,
		},
		TotalStatistics: Stats{
			Size: item.Size,
		},
		Error: nil,
	}
}

func newProgressError(err error) Progress {
	return Progress{
		Error: err,
	}
}

// TransferSpeed returns the average transfer speed in bytes per second
func (p Progress) TransferSpeed() uint64 {
	return uint64(float64(p.CurrentItemStats.Transferred) / time.Since(p.Timer).Seconds())
}
