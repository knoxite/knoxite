package main

import (
	"errors"
	"fmt"

	"github.com/knoxite/knoxite"
)

// CmdLs describes the command
type CmdLs struct {
	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("ls",
		"list files",
		"The ls command lists all files stored in a snapshot",
		&CmdLs{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdLs) Usage() string {
	return "SNAPSHOT-ID" // "[DIR/FILE] [...]"
}

// Execute this command
func (cmd CmdLs) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return errors.New(TSpecifyRepoLocation)
	}

	repository, err := knoxite.OpenRepository(cmd.global.Repo, cmd.global.Password)
	if err == nil {
		tab := NewTable([]string{"Perms", "User", "Group", "Size", "ModTime", "Name"}, []int64{-10, -5, -5, 12, -19, -48}, "No files found.")

		_, snapshot, ferr := repository.FindSnapshot(args[0])
		if ferr != nil {
			return ferr
		}

		for _, archive := range snapshot.Items {
			tab.Rows = append(tab.Rows, []interface{}{archive.Mode, "user", "group", knoxite.SizeToString(archive.Size), archive.ModTime.Format(timeFormat), archive.Path})
		}

		tab.Print()
	}

	return err
}
