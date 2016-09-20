package main

import (
	"errors"
	"fmt"
	"os/user"
	"strconv"

	"github.com/knoxite/knoxite"
	"github.com/muesli/gotable"
)

const timeFormat = "2006-01-02 15:04:05"

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

	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err == nil {
		tab := gotable.NewTable([]string{"Perms", "User", "Group", "Size", "ModTime", "Name"},
			[]int64{-10, -8, -5, 12, -19, -48},
			"No files found.")

		_, snapshot, ferr := repository.FindSnapshot(args[0])
		if ferr != nil {
			return ferr
		}

		for _, archive := range snapshot.Items {
			username := strconv.FormatInt(int64(archive.UID), 10)
			u, uerr := user.LookupId(username)
			if uerr == nil {
				username = u.Username
			}
			groupname := strconv.FormatInt(int64(archive.GID), 10)
			tab.AppendRow([]interface{}{
				archive.Mode,
				username,
				groupname,
				knoxite.SizeToString(archive.Size),
				archive.ModTime.Format(timeFormat),
				archive.Path})
		}

		tab.Print()
	}

	return err
}
