/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/shutdown2"
	"github.com/muesli/goprogressbar"

	"github.com/knoxite/knoxite"
)

// Error declarations
var (
	ErrRedundancyAmount = errors.New("failure tolerance can't be equal or higher as the number of storage backends")
)

// CmdStore describes the command
type CmdStore struct {
	Description      string `short:"d" long:"desc"        description:"a description or comment for this snapshot"`
	Compression      string `short:"c" long:"compression" description:"compression algo to use: none (default), gzip"`
	Encryption       string `short:"e" long:"encryption"  description:"encryption algo to use: aes (default), none"`
	FailureTolerance uint   `short:"t" long:"tolerance"   description:"failure tolerance against n backend failures"`

	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("store",
		"store file/directory",
		"The store command creates a snapshot of a file or directory",
		&CmdStore{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

func (cmd CmdStore) store(repository *knoxite.Repository, chunkIndex *knoxite.ChunkIndex, snapshot *knoxite.Snapshot, targets []string) error {
	// we want to be notified during the first phase of a shutdown
	cancel := shutdown.First()

	fmt.Println()
	overallProgressBar := goprogressbar.NewProgressBar("Overall Progress", 0, 0, 60)
	wd, gerr := os.Getwd()
	if gerr != nil {
		return gerr
	}

	if uint(len(repository.Backend.Backends))-cmd.FailureTolerance <= 0 {
		return ErrRedundancyAmount
	}

	progress := snapshot.Add(wd, targets, *repository, chunkIndex,
		strings.ToLower(cmd.Compression) == "gzip", strings.ToLower(cmd.Encryption) != "none",
		uint(len(repository.Backend.Backends))-cmd.FailureTolerance, cmd.FailureTolerance)

	fileProgressBar := goprogressbar.NewProgressBar("", 0, 0, 60)
	lastPath := ""

	for p := range progress {
		select {
		case n := <-cancel:
			fmt.Println("Aborting...")
			close(n)
			return nil

		default:
			if p.Error != nil {
				fmt.Println()
				return p.Error
			}
			if p.Path != lastPath && lastPath != "" {
				fmt.Println()
			}
			fileProgressBar.Total = int64(p.Size)
			fileProgressBar.Current = int64(p.Transferred)
			fileProgressBar.RightAlignedText = fmt.Sprintf("%s / %s  %s/s",
				knoxite.SizeToString(uint64(fileProgressBar.Current)),
				knoxite.SizeToString(uint64(fileProgressBar.Total)),
				knoxite.SizeToString(p.TransferSpeed()))

			overallProgressBar.Total = int64(p.Statistics.Size)
			overallProgressBar.Current = int64(p.Statistics.Transferred)
			overallProgressBar.RightAlignedText = fmt.Sprintf("%s / %s",
				knoxite.SizeToString(uint64(overallProgressBar.Current)),
				knoxite.SizeToString(uint64(overallProgressBar.Total)))

			if p.Path != lastPath {
				lastPath = p.Path
				fileProgressBar.Text = p.Path
			}

			goprogressbar.MoveCursorUp(1)
			fileProgressBar.Print()
			goprogressbar.MoveCursorDown(1)
			overallProgressBar.Print()
		}
	}

	fmt.Printf("\nSnapshot %s created: %s\n", snapshot.ID, snapshot.Stats.String())
	return nil
}

// Usage describes this command's usage help-text
func (cmd CmdStore) Usage() string {
	return "VOLUME-ID DIR/FILE [DIR/FILE] [...]"
}

// Execute this command
func (cmd CmdStore) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return ErrMissingRepoLocation
	}

	targets := []string{}
	for _, target := range args[1:] {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// filter here? exclude/include?

	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return err
	}
	volume, err := repository.FindVolume(args[0])
	if err != nil {
		return err
	}
	snapshot, err := knoxite.NewSnapshot(cmd.Description)
	if err != nil {
		return err
	}
	chunkIndex, err := knoxite.OpenChunkIndex(&repository)
	if err != nil {
		return err
	}
	// release the shutdown lock
	lock()

	err = cmd.store(&repository, &chunkIndex, &snapshot, targets)
	if err != nil {
		return err
	}

	// acquire another shutdown lock. we don't want these next calls to be interrupted
	lock = shutdown.Lock()
	if lock == nil {
		return nil
	}
	defer lock()

	err = snapshot.Save(&repository)
	if err != nil {
		return err
	}
	err = volume.AddSnapshot(snapshot.ID)
	if err != nil {
		return err
	}
	err = chunkIndex.Save(&repository)
	if err != nil {
		return err
	}
	return repository.Save()
}
