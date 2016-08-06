package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/knoxite/knoxite"
)

// CmdRepository describes the command
type CmdRepository struct {
	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("repo",
		"manage repository",
		"The repo command manages repositories",
		&CmdRepository{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdRepository) Usage() string {
	return "[init|add|cat]"
}

// Execute this command
func (cmd CmdRepository) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return errors.New(TSpecifyRepoLocation)
	}

	switch args[0] {
	case "init":
		return cmd.init()
	case "add":
		if len(args) < 2 {
			return fmt.Errorf(TWrongNumArgs, cmd.Usage())
		}
		return cmd.add(args[1])
	case "cat":
		return cmd.cat()
	}

	return nil
}

func (cmd CmdRepository) init() error {
	/*	username := ""
		user, err := user.Current()
		if err == nil {
			username = user.Username
		}
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}*/

	_, err := newRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return fmt.Errorf("Creating repository at %s failed: %v", cmd.global.Repo, err)
	}

	fmt.Printf("Created new repository at %s\n", cmd.global.Repo)
	return nil
}

func (cmd CmdRepository) add(url string) error {
	r, err := openRepository(cmd.global.Repo, cmd.global.Password)
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

func (cmd CmdRepository) cat() error {
	r, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, r.RawJSON, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(out.Bytes()))
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
		password, err = readPasswordTwice("Enter password:", "Confirm password:")
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
		return pw, errors.New("Passwords did not match")
	}

	return pw, nil
}
