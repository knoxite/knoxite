package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knoxite/knoxite"
)

// CmdStore describes the command
type CmdStore struct {
	Description string `short:"d" long:"desc"        description:"a description or comment for this snapshot"`
	Compression string `short:"c" long:"compression" description:"compression algo to use: none (default), gzip"`
	Encryption  string `short:"e" long:"encryption"  description:"encryption algo to use: aes (default), none"`

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
		return errors.New(TSpecifyRepoLocation)
	}

	targets := []string{}
	for _, target := range args[1:] {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// filter here? exclude/include?

	repository, err := knoxite.OpenRepository(cmd.global.Repo, cmd.global.Password)
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

	for _, target := range targets {
		wd, gerr := os.Getwd()
		if gerr != nil {
			return gerr
		}
		progress, serr := snapshot.Add(wd, target, repository, strings.ToLower(cmd.Compression) == "gzip", strings.ToLower(cmd.Encryption) != "none")
		if serr != nil {
			return serr
		}

		for p := range progress {
			fmt.Printf("\033[2K\r%s - [%s]", p.Stats.String(), p.Path)
		}
	}

	fmt.Printf("\nSnapshot %s created: %s\n", snapshot.ID, snapshot.Stats.String())

	err = snapshot.Save(&repository)
	if err != nil {
		return err
	}
	err = volume.AddSnapshot(snapshot.ID)
	if err != nil {
		return err
	}
	return repository.Save()
}
