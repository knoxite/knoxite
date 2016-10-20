package knoxite

// Progress contains stats and current path
type Progress struct {
	Path        string
	Size        uint64
	StorageSize uint64
	Statistics  Stats
	Error       error
}

func newProgress(item *ItemData) Progress {
	return Progress{
		Path:        item.Path,
		Size:        item.Size,
		StorageSize: item.StorageSize,
		Statistics:  Stats{},
		Error:       nil,
	}
}

func newProgressError(err error) Progress {
	return Progress{
		Error: err,
	}
}
