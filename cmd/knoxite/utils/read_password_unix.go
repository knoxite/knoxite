//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package utils

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"golang.org/x/term"
)

func ReadPassword(prompt string) (string, error) {
	var tty io.WriteCloser
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		tty = os.Stdout
	} else {
		defer tty.Close()
	}

	fmt.Fprint(tty, prompt+" ")
	buf, err := term.ReadPassword(syscall.Stdin)
	fmt.Fprintln(tty)

	return string(buf), err
}
