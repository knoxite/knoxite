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
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/klauspost/shutdown2"
)

// Translations
const (
	TWrongNumArgs        = "Wrong number of arguments, usage: %s"
	TUnknownCommand      = "Unknown command, usage: %s"
	TSpecifyRepoLocation = "Please specify repository location (-r)"
)

// Error declarations
var (
	ErrMissingRepoLocation = errors.New(TSpecifyRepoLocation)
)

// GlobalOptions holds all those options that can be set for every command
type GlobalOptions struct {
	Repo     string `short:"r" long:"repo"     description:"Repository directory to backup to/restore from"`
	Password string `short:"p" long:"password" description:"Password to use for data encryption"`
}

var (
	globalOpts = GlobalOptions{}
	parser     = flags.NewParser(&globalOpts, flags.HelpFlag|flags.PassDoubleDash)
)

func main() {
	shutdown.OnSignal(0, os.Interrupt, syscall.SIGTERM)
	// quiet shutdown logger
	shutdown.Logger = shutdown.LogPrinter(log.New(ioutil.Discard, "", log.LstdFlags))
	// shutdown.SetTimeout(0)

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	// fmt.Println("Exiting.")
}
