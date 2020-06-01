/*
 * knoxite
 *     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"math"
	"math/rand"
)

func VerifyRepo(repository Repository, Percentage int) (prog chan Progress, err error) {
	prog = make(chan Progress)

	go func() {

		archiveToSnapshot := make(map[string]*Snapshot)

		for _, volume := range repository.Volumes {
			for _, snapshotHash := range volume.Snapshots {
				_, snapshot, err := repository.FindSnapshot(snapshotHash)
				if err != nil {
					prog <- newProgressError(err)
				}

				for archiveHash := range snapshot.Archives {
					archiveToSnapshot[archiveHash] = snapshot
				}

			}
		}

		archives := make([]string, 0)
		for archiveHash := range archiveToSnapshot {
			archives = append(archives, archiveHash)
		}

		// get all keys of the snapshot Archives

		if Percentage > 100 {
			Percentage = 100
		} else if Percentage < 0 {
			Percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*Percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx := rand.Intn(len(archives))
			selectedArchives[archives[idx]] = true
		}

		for archiveKey := range selectedArchives {
			snapshot := archiveToSnapshot[archiveKey]
			p := newProgress(snapshot.Archives[archiveKey])
			prog <- p

			err := VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				prog <- newProgressError(err)
			}

			p.CurrentItemStats.Transferred += uint64((*snapshot.Archives[archiveKey]).Size)
			prog <- p

		}
		close(prog)
	}()

	return prog, nil
}

func VerifyVolume(repository Repository, volumeId string, Percentage int) (prog chan Progress, err error) {
	prog = make(chan Progress)

	go func() {
		volume, ferr := repository.FindVolume(volumeId)
		if ferr != nil {
			prog <- newProgressError(ferr)
		}

		archiveToSnapshot := make(map[string]*Snapshot)

		for _, snapshotHash := range volume.Snapshots {
			_, snapshot, err := repository.FindSnapshot(snapshotHash)
			if err != nil {
				prog <- newProgressError(err)
			}

			for archiveHash := range snapshot.Archives {
				archiveToSnapshot[archiveHash] = snapshot
			}

		}

		archives := make([]string, 0)
		for archiveHash := range archiveToSnapshot {
			archives = append(archives, archiveHash)
		}

		// get all keys of the snapshot Archives

		if Percentage > 100 {
			Percentage = 100
		} else if Percentage < 0 {
			Percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*Percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx := rand.Intn(len(archives))
			selectedArchives[archives[idx]] = true
		}

		for archiveKey := range selectedArchives {
			snapshot := archiveToSnapshot[archiveKey]
			p := newProgress(snapshot.Archives[archiveKey])
			prog <- p

			err := VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				prog <- newProgressError(err)
			}

			p.CurrentItemStats.Transferred += uint64((*snapshot.Archives[archiveKey]).Size)
			prog <- p

		}
		close(prog)
	}()

	return prog, nil
}

func VerifySnapshot(repository Repository, snapshotId string, Percentage int) (prog chan Progress, err error) {
	prog = make(chan Progress)

	go func() {
		_, snapshot, ferr := repository.FindSnapshot(snapshotId)
		if ferr != nil {
			prog <- newProgressError(ferr)
		}

		// get all keys of the snapshot Archives
		archives := make([]string, 0, len(snapshot.Archives))
		for key := range snapshot.Archives {
			archives = append(archives, key)
		}

		if Percentage > 100 {
			Percentage = 100
		} else if Percentage < 0 {
			Percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*Percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx := rand.Intn(len(archives))
			selectedArchives[archives[idx]] = true
		}

		for archiveKey := range selectedArchives {
			p := newProgress(snapshot.Archives[archiveKey])
			prog <- p

			err := VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				prog <- newProgressError(err)
			}

			p.CurrentItemStats.Transferred += uint64((*snapshot.Archives[archiveKey]).Size)
			prog <- p

		}
		close(prog)
	}()

	return prog, nil
}

func VerifyArchive(repository Repository, arc Archive) error {
	if arc.Type == File {
		parts := uint(len(arc.Chunks))
		for i := uint(0); i < parts; i++ {
			idx, erri := arc.IndexOfChunk(i)
			if erri != nil {
				return erri
			}

			chunk := arc.Chunks[idx]
			_, errc := loadChunk(repository, arc, chunk)
			if errc != nil {
				return errc
			}

		}
		return nil
	}
	return nil

}
