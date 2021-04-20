/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020,      Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */

package main

import (
	"encoding/json"
	"fmt"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/muesli/gotable"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/cmd/knoxite/utils"
)

var (
	repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "manage repository",
		Long:  `The repo command manages repositories`,
		RunE:  nil,
	}
	repoInitCmd = &cobra.Command{
		Use:   "init",
		Short: "initialize a new repository",
		Long:  `The init command initializes a new repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRepoInit()
		},
	}
	repoChangePasswordCmd = &cobra.Command{
		Use:   "passwd",
		Short: "changes the password of a repository",
		Long:  `The passwd command changes the password of a repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRepoChangePassword()
		},
	}
	repoCatCmd = &cobra.Command{
		Use:   "cat",
		Short: "display repository information as JSON",
		Long:  `The cat command displays the internal repository information as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRepoCat()
		},
	}
	repoInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "display repository information",
		Long:  `The info command displays the repository status & information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRepoInfo()
		},
	}
	repoAddCmd = &cobra.Command{
		Use:   "add <url>",
		Short: "add another storage backend to a repository",
		Long:  `The add command adds another storage backend to a repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("add needs a URL to be added")
			}
			return executeRepoAdd(args[0])
		},
	}
	repoPackCmd = &cobra.Command{
		Use:   "pack",
		Short: "pack repository and release redundant data",
		Long:  `The pack command deletes all unused data chunks from storage`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRepoPack()
		},
	}
	setURLCmd = &cobra.Command{
		Use:   "set-url <new-url>",
		Short: "set a new URL for the repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("New URL is needed for changing URL")
			}
			return executeRepoChangeLocation(args[0])
		},
	}
)

func init() {
	repoCmd.AddCommand(repoInitCmd)
	repoCmd.AddCommand(repoChangePasswordCmd)
	repoCmd.AddCommand(repoCatCmd)
	repoCmd.AddCommand(repoInfoCmd)
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoPackCmd)
	repoCmd.AddCommand(setURLCmd)
	RootCmd.AddCommand(repoCmd)
}

func executeRepoInit() error {
	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	defer lock()

	r, err := newRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return fmt.Errorf("Creating repository at %s failed: %v", globalOpts.Repo, err)
	}

	fmt.Printf("Created new repository at %s\n", (*r.BackendManager().Backends[0]).Location())
	return nil
}

func executeRepoChangePassword() error {
	r, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	password, err := utils.ReadPasswordTwice("Enter new password:", "Confirm password:")
	if err != nil {
		return err
	}

	err = r.ChangePassword(password)
	if err != nil {
		return err
	}

	fmt.Printf("Changed password successfully\n")
	return nil
}

func executeRepoAdd(url string) error {
	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	defer lock()

	r, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	backend, err := knoxite.BackendFromURL(url)
	if err != nil {
		return err
	}

	err = backend.InitRepository()
	if err != nil {
		return err
	}

	r.BackendManager().AddBackend(&backend)

	err = r.Save()
	if err != nil {
		return err
	}
	fmt.Printf("Added %s to repository\n", backend.Location())
	return nil
}

func executeRepoCat() error {
	r, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	json, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", json)
	return nil
}

func executeRepoPack() error {
	r, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	index, err := knoxite.OpenChunkIndex(&r)
	if err != nil {
		return err
	}

	freedSize, err := index.Pack(&r)
	if err != nil {
		return err
	}

	err = index.Save(&r)
	if err != nil {
		return err
	}

	fmt.Printf("Freed storage space: %s\n", knoxite.SizeToString(freedSize))
	return nil
}

func executeRepoInfo() error {
	r, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	tab := gotable.NewTable([]string{"Storage URL", "Available Space"},
		[]int64{-48, 15},
		"No backends found.")

	for _, be := range r.BackendManager().Backends {
		space, _ := (*be).AvailableSpace()
		tab.AppendRow([]interface{}{
			(*be).Location(),
			knoxite.SizeToString(space)})
	}

	_ = tab.Print()
	return nil
}

func executeRepoChangeLocation(newLocation string) error {
	r, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	err = r.ChangeLocation(globalOpts.Repo, newLocation)
	if err != nil {
		return err
	}

	fmt.Printf("Location successfully changed to \"%s\"\n", newLocation)
	return nil
}

func openRepository(path, password string) (knoxite.Repository, error) {
	if password == "" {
		var err error
		password, err = utils.ReadPassword("Enter password:")
		if err != nil {
			return knoxite.Repository{}, err
		}
	}

	return knoxite.OpenRepository(path, password)
}

func newRepository(path, password string) (knoxite.Repository, error) {
	if password == "" {
		var err error
		password, err = utils.ReadPasswordTwice("Enter a password to encrypt this repository with:", "Confirm password:")
		if err != nil {
			return knoxite.Repository{}, err
		}
	}

	return knoxite.NewRepository(path, password)
}
