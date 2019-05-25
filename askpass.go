// askpass.go -- Interactive password prompt
//
// (c) 2016 Sudhi Herle <sudhi@herle.net>
// (c) 2019 Saeed Rasooli <saeed.gnu@gmail.com>
//
// Placed in the Public Domain
//
// This software does not come with any express or implied
// warranty; it is provided "as is". No claim  is made to its
// suitability for any purpose.

package askpass

import (
	"fmt"
	"os"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// Askpass prompts user for an interactive password.
// If verify is true, confirm a second time.
// Mistakes during confirmation cause the process to restart upto a
// maximum of 2 times.
func Askpass(prompt string, verify bool) (string, error) {
	var fd int
	var out *os.File
	if terminal.IsTerminal(syscall.Stdin) {
		fd = syscall.Stdin
		out = os.Stdout
	} else {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return "", errors.Wrap(err, "error allocating terminal")
		}
		defer tty.Close()
		fd = int(tty.Fd())
	}
	if terminal.IsTerminal(syscall.Stdout) {
		out = os.Stdout
	} else if terminal.IsTerminal(syscall.Stderr) {
		out = os.Stderr
	} else {
		return "", fmt.Errorf("neither stdout not stderr is a terminal")
		// TODO: use askpass programs
	}

	for i := 0; i < 2; i++ {
		out.WriteString(prompt + ": ")
		pw1, err := terminal.ReadPassword(fd)
		out.WriteString("\n")
		if err != nil {
			return "", err
		}
		if !verify {
			return string(pw1), nil
		}

		out.WriteString(prompt + " again: ")
		pw2, err := terminal.ReadPassword(fd)
		out.WriteString("\n")
		if err != nil {
			return "", err
		}

		a := string(pw1)
		b := string(pw2)
		if a == b {
			return a, nil
		}

		out.WriteString("** password mismatch; try again ..\n")
	}

	return "", fmt.Errorf("Too many tries getting password")
}

// vim: ft=go:sw=8:ts=8:noexpandtab:tw=98:
