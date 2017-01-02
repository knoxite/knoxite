/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall"

	"github.com/klauspost/shutdown2"
	"github.com/muesli/gotable"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	knoxite "github.com/knoxite/knoxite/lib"
)

// Error declarations
var (
	ErrPasswordMismatch = errors.New("Passwords did not match")

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
)

func init() {
	repoCmd.AddCommand(repoInitCmd)
	repoCmd.AddCommand(repoCatCmd)
	repoCmd.AddCommand(repoInfoCmd)
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoPackCmd)
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

	fmt.Printf("Created new repository at %s\n", (*r.Backend.Backends[0]).Location())
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
	r.Backend.AddBackend(&backend)

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

	for _, be := range r.Backend.Backends {
		space, _ := (*be).AvailableSpace()
		tab.AppendRow([]interface{}{
			(*be).Location(),
			knoxite.SizeToString(space)})
	}

	tab.Print()
	return nil
}

func openRepository(path, password string) (knoxite.Repository, error) {
	if password == "" {
		var err error
		password, err = readPassword("Enter password:")
		if err != nil {
			return knoxite.Repository{}, err
		}
	}

	return knoxite.OpenRepository(path, password)
}

func newRepository(path, password string) (knoxite.Repository, error) {
	if password == "" {
		var err error
		password, err = readPasswordTwice("Enter a password to encrypt this repository with:", "Confirm password:")
		if err != nil {
			return knoxite.Repository{}, err
		}
	}

	return knoxite.NewRepository(path, password)
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt + " ")
	buf, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	return string(buf), err
}

func readPasswordTwice(prompt, promptConfirm string) (string, error) {
	pw, err := readPassword(prompt)
	if err != nil {
		return pw, err
	}

	pwconfirm, err := readPassword(promptConfirm)
	if err != nil {
		return pw, err
	}
	if pw != pwconfirm {
		return pw, ErrPasswordMismatch
	}

	return pw, nil
}
