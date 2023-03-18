/*
 * knoxite
 *     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"crypto/rand"
	"math"
	"math/big"
)

func VerifyRepo(repository Repository, percentage int) (<-chan Progress, error) {
	prog := make(chan Progress)

	go func() {
		defer close(prog)
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

		// get all keys of the snapshot Archives
		archives := make([]string, 0)
		for archiveHash := range archiveToSnapshot {
			archives = append(archives, archiveHash)
		}

		if percentage > 100 {
			percentage = 100
		} else if percentage < 0 {
			percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(archives))))
			if err != nil {
				prog <- newProgressError(err)
			}
			selectedArchives[archives[idx.Int64()]] = true
		}

		for archiveKey := range selectedArchives {
			snapshot := archiveToSnapshot[archiveKey]
			p := newProgress(snapshot.Archives[archiveKey])
			prog <- p

			err := VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				prog <- newProgressError(err)
			}

			p.CurrentItemStats.Transferred += (*snapshot.Archives[archiveKey]).Size
			prog <- p
		}
	}()

	return prog, nil
}

func VerifyVolume(repository Repository, volumeId string, percentage int) (<-chan Progress, error) {
	prog := make(chan Progress)

	go func() {
		defer close(prog)
		volume, err := repository.FindVolume(volumeId)
		if err != nil {
			prog <- newProgressError(err)
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

		// get all keys of the snapshot Archives
		archives := make([]string, 0)
		for archiveHash := range archiveToSnapshot {
			archives = append(archives, archiveHash)
		}

		if percentage > 100 {
			percentage = 100
		} else if percentage < 0 {
			percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(archives))))
			if err != nil {
				prog <- newProgressError(err)
			}
			selectedArchives[archives[idx.Int64()]] = true
		}

		for archiveKey := range selectedArchives {
			snapshot := archiveToSnapshot[archiveKey]
			p := newProgress(snapshot.Archives[archiveKey])
			prog <- p

			err := VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				prog <- newProgressError(err)
			}

			p.CurrentItemStats.Transferred += (*snapshot.Archives[archiveKey]).Size
			prog <- p
		}
	}()

	return prog, nil
}

func VerifySnapshot(repository Repository, snapshotId string, percentage int) (<-chan Progress, error) {
	prog := make(chan Progress)

	go func() {
		defer close(prog)
		_, snapshot, err := repository.FindSnapshot(snapshotId)
		if err != nil {
			prog <- newProgressError(err)
		}

		// get all keys of the snapshot Archives
		archives := make([]string, 0, len(snapshot.Archives))
		for key := range snapshot.Archives {
			archives = append(archives, key)
		}

		if percentage > 100 {
			percentage = 100
		} else if percentage < 0 {
			percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(archives))))
			if err != nil {
				prog <- newProgressError(err)
			}
			selectedArchives[archives[idx.Int64()]] = true
		}

		for archiveKey := range selectedArchives {
			p := newProgress(snapshot.Archives[archiveKey])
			prog <- p

			err := VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				prog <- newProgressError(err)
			}

			p.CurrentItemStats.Transferred += (*snapshot.Archives[archiveKey]).Size
			prog <- p
		}
	}()

	return prog, nil
}

func VerifyArchive(repository Repository, arc Archive) error {
	if arc.Type != File {
		return nil
	}

	parts := uint(len(arc.Chunks))
	for i := uint(0); i < parts; i++ {
		idx, err := arc.IndexOfChunk(i)
		if err != nil {
			return err
		}

		chunk := arc.Chunks[idx]
		_, err = loadChunk(repository, arc, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}
