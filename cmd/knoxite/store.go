/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	shutdown "github.com/klauspost/shutdown2"
	"github.com/muesli/goprogressbar"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/knoxite/knoxite"
)

// Error declarations
var (
	ErrRedundancyAmount   = errors.New("failure tolerance can't be equal or higher as the number of storage backends")
	ErrEncryptionUnknown  = errors.New("unknown encryption format")
	ErrCompressionUnknown = errors.New("unknown compression format")
)

// StoreOptions holds all the options that can be set for the 'store' command
type StoreOptions struct {
	Description      string
	Compression      string
	Encryption       string
	FailureTolerance uint
	Excludes         []string
	ExcludeExternalSymlinks bool
}

var (
	storeOpts = StoreOptions{}

	storeCmd = &cobra.Command{
		Use:   "store <volume> <dir/file> [...]",
		Short: "store files/directories",
		Long:  `The store command creates a snapshot of a file or directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("store needs to know which volume to create a snapshot in")
			}
			if len(args) < 2 {
				return fmt.Errorf("store needs to know which files and/or directories to work on")
			}
			return executeStore(args[0], args[1:], storeOpts)
		},
	}
)

func initStoreFlags(f func() *pflag.FlagSet) {
	f().StringVarP(&storeOpts.Description, "desc", "d", "", "a description or comment for this volume")
	f().StringVarP(&storeOpts.Compression, "compression", "c", "", "compression algo to use: none (default), flate, gzip, lzma, zlib, zstd")
	f().StringVarP(&storeOpts.Encryption, "encryption", "e", "", "encryption algo to use: aes (default), none")
	f().UintVarP(&storeOpts.FailureTolerance, "tolerance", "t", 0, "failure tolerance against n backend failures")
	f().StringArrayVarP(&storeOpts.Excludes, "excludes", "x", []string{}, "list of excludes")
	f().BoolVar(&storeOpts.ExcludeExternalSymlinks, "exclude-external-symlinks", false, "Exclude Symlinks not included in the snapshot")
}

func init() {
	initStoreFlags(storeCmd.Flags)
	RootCmd.AddCommand(storeCmd)
}

func store(repository *knoxite.Repository, chunkIndex *knoxite.ChunkIndex, snapshot *knoxite.Snapshot, targets []string, opts StoreOptions) error {
	// we want to be notified during the first phase of a shutdown
	cancel := shutdown.First()

	wd, gerr := os.Getwd()
	if gerr != nil {
		return gerr
	}

	if uint(len(repository.BackendManager().Backends))-opts.FailureTolerance <= 0 {
		return ErrRedundancyAmount
	}
	compression, err := CompressionTypeFromString(opts.Compression)
	if err != nil {
		return err
	}
	encryption, err := EncryptionTypeFromString(opts.Encryption)
	if err != nil {
		return err
	}

	startTime := time.Now()
	progress := snapshot.Add(wd, targets, opts.Excludes, *repository, chunkIndex,
		compression, encryption,
		uint(len(repository.BackendManager().Backends))-opts.FailureTolerance, opts.FailureTolerance, opts.ExcludeExternalSymlinks)

	fileProgressBar := &goprogressbar.ProgressBar{Width: 40}
	overallProgressBar := &goprogressbar.ProgressBar{
		Text:  fmt.Sprintf("%d of %d total", 0, 0),
		Width: 60,
		PrependTextFunc: func(p *goprogressbar.ProgressBar) string {
			return fmt.Sprintf("%s/s",
				knoxite.SizeToString(uint64(float64(p.Current)/time.Since(startTime).Seconds())))
		},
	}

	pb := goprogressbar.MultiProgressBar{}
	pb.AddProgressBar(fileProgressBar)
	pb.AddProgressBar(overallProgressBar)
	lastPath := ""

	items := int64(1)
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
				items++
				fmt.Println()
			}
			fileProgressBar.Total = int64(p.CurrentItemStats.Size)
			fileProgressBar.Current = int64(p.CurrentItemStats.Transferred)
			fileProgressBar.PrependText = fmt.Sprintf("%s  %s/s",
				knoxite.SizeToString(uint64(fileProgressBar.Current)),
				knoxite.SizeToString(p.TransferSpeed()))

			overallProgressBar.Total = int64(p.TotalStatistics.Size)
			overallProgressBar.Current = int64(p.TotalStatistics.Transferred)
			overallProgressBar.Text = fmt.Sprintf("%s / %s (%s of %s)",
				knoxite.SizeToString(uint64(overallProgressBar.Current)),
				knoxite.SizeToString(uint64(overallProgressBar.Total)),
				humanize.Comma(items),
				humanize.Comma(int64(p.TotalStatistics.Files+p.TotalStatistics.Dirs+p.TotalStatistics.SymLinks)))

			if p.Path != lastPath {
				lastPath = p.Path
				fileProgressBar.Text = p.Path
			}

			pb.LazyPrint()
		}
	}

	fmt.Printf("\nSnapshot %s created: %s\n", snapshot.ID, snapshot.Stats.String())
	return nil
}

func executeStore(volumeID string, args []string, opts StoreOptions) error {
	targets := []string{}
	for _, target := range args {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	volume, err := repository.FindVolume(volumeID)
	if err != nil {
		return err
	}
	snapshot, err := knoxite.NewSnapshot(opts.Description)
	if err != nil {
		return err
	}
	chunkIndex, err := knoxite.OpenChunkIndex(&repository)
	if err != nil {
		return err
	}
	// release the shutdown lock
	lock()

	err = store(&repository, &chunkIndex, snapshot, targets, opts)
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

// CompressionTypeFromString returns the compression type from a user-specified string
func CompressionTypeFromString(s string) (uint16, error) {
	switch strings.ToLower(s) {
	case "":
		// default is none
		fallthrough
	case "none":
		return knoxite.CompressionNone, nil
	case "flate":
		return knoxite.CompressionFlate, nil
	case "gzip":
		return knoxite.CompressionGZip, nil
	case "lzma":
		return knoxite.CompressionLZMA, nil
	case "zlib":
		return knoxite.CompressionZlib, nil
	case "zstd":
		return knoxite.CompressionZstd, nil
	}

	return 0, ErrCompressionUnknown
}

// CompressionText returns a user-friendly string indicating the compression algo that was used
func CompressionText(enum int) string {
	switch enum {
	case knoxite.CompressionNone:
		return "none"
	case knoxite.CompressionFlate:
		return "Flate"
	case knoxite.CompressionGZip:
		return "GZip"
	case knoxite.CompressionLZMA:
		return "LZMA"
	case knoxite.CompressionZlib:
		return "zlib"
	case knoxite.CompressionZstd:
		return "zstd"
	}

	return "unknown"
}

// EncryptionTypeFromString returns the encryption type from a user-specified string
func EncryptionTypeFromString(s string) (uint16, error) {
	switch strings.ToLower(s) {
	case "":
		// default is AES
		fallthrough
	case "aes":
		return knoxite.EncryptionAES, nil
	case "none":
		return knoxite.EncryptionNone, nil
	}

	return 0, ErrEncryptionUnknown
}

// EncryptionText returns a user-friendly string indicating the encryption algo that was used
func EncryptionText(enum int) string {
	switch enum {
	case knoxite.EncryptionNone:
		return "none"
	case knoxite.EncryptionAES:
		return "AES"
	}

	return "unknown"
}
