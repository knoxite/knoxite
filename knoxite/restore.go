package main

import (
	"errors"
	"fmt"

	"github.com/knoxite/knoxite"
)

// CmdRestore describes the command
type CmdRestore struct {
	Target string `short:"t" long:"target" description:"Directory to restore to"`

	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("restore",
		"restore a snapshot",
		"The restore command restores a snapshot to a directory",
		&CmdRestore{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdRestore) Usage() string {
	return "SNAPSHOT-ID"
}

// Execute this command
func (cmd CmdRestore) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return errors.New(TSpecifyRepoLocation)
	}
	if cmd.Target == "" {
		return errors.New("please specify a directory to restore to (--target)")
	}

	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err == nil {
		_, snapshot, ferr := repository.FindSnapshot(args[0])
		if ferr != nil {
			return ferr
		}

		stats, derr := knoxite.DecodeSnapshot(repository, *snapshot, cmd.Target)
		if derr != nil {
			return derr
		}
		fmt.Println("Restore done:", stats.String())
		return nil
	}

	return err
}
