/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"os/user"
	"strconv"
	"time"

	"github.com/muesli/gotable"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
)

const timeFormat = "2006-01-02 15:04:05"

var (
	lsCmd = &cobra.Command{
		Use:   "ls <snapshot>",
		Short: "list files",
		Long:  `The ls command lists all files stored in a snapshot`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("ls needs a snapshot ID")
			}
			return executeLs(args[0])
		},
	}
)

func init() {
	RootCmd.AddCommand(lsCmd)
}

func executeLs(snapshotID string) error {
	logger.Info("Opening repository")
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	logger.Info("Opened repository")

	logger.Debug("Initialising new gotable for output")
	tab := gotable.NewTable([]string{"Perms", "User", "Group", "Size", "ModTime", "Name"},
		[]int64{-10, -8, -5, 12, -19, -48},
		"No files found.")

	logger.Info("Finding snapshot " + snapshotID)
	_, snapshot, err := repository.FindSnapshot(snapshotID)
	if err != nil {
		return err
	}
	logger.Info("Found snapshot " + snapshot.Description)

	logger.Debug("Iterating over archives to print details")
	for _, archive := range snapshot.Archives {
		username := strconv.FormatInt(int64(archive.UID), 10)

		logger.Infof("Looking up OS username with archive's UID %d", archive.UID)
		u, err := user.LookupId(username)
		if err == nil {
			logger.Infof("Username found: %s", u.Username)
			username = u.Username
		} else {
			logger.Warn("Looking up username failed. Using default value.")
		}

		groupname := strconv.FormatInt(int64(archive.GID), 10)
		tab.AppendRow([]interface{}{
			archive.Mode,
			username,
			groupname,
			knoxite.SizeToString(archive.Size),
			time.Unix(archive.ModTime, 0).Format(timeFormat),
			archive.Path})
	}

	logger.Debug("Printing ls output")
	err = tab.Print()
	if err != nil {
		return err
	}
	logger.Debug("Printed ls output")

	return nil
}
