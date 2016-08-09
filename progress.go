package knoxite

// Progress contains stats and current path
type Progress struct {
	Path        string
	Size        uint64
	StorageSize uint64
	Statistics  Stats
}

func newProgress(item *ItemData) Progress {
	return Progress{
		Path:        item.Path,
		Size:        item.Size,
		StorageSize: item.StorageSize,
	}
}
