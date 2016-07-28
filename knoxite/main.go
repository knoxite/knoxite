package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
)

const (
	TWrongNumArgs        = "wrong number of arguments, Usage: %s"
	TSpecifyRepoLocation = "Please specify repository location (-r)"
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

func handleSignals() {
	// Wait for signals
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	for s := range ch {
		fmt.Println("Got signal:", s)

		switch s {
		case syscall.SIGHUP:
			fallthrough
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGINT:
			return
		}
	}
}

func main() {
	go func() {
		handleSignals()
	}()

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	//	fmt.Println("Exiting.")
}
