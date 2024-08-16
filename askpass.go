// askpass.go -- Interactive password prompt
//
// (c) 2016 Sudhi Herle <sudhi@herle.net>
// (c) 2019 Saeed Rasooli <saeed.gnu@gmail.com>
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 2 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/lgpl.txt>.
// Also avalable in /usr/share/common-licenses/LGPL on Debian systems
// or /usr/share/licenses/common/LGPL/license.txt on ArchLinux

package askpass

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/term"
)

func askpassOnce(prompt string, out *os.File) (string, error) {
	askpassPath := getAskpassBinaryPath()

	if askpassPath != "" {
		stdout, stderr, exitCode := RunCommand3(askpassPath, prompt)
		if exitCode != 0 {
			if stderr != "" {
				return "", fmt.Errorf("error from %v:\n%v", askpassPath, stderr)
			}
		}
		if stdout == "" {
			return "", fmt.Errorf("entered empty password")
		}
		return stdout, nil
	}
	var fd uintptr
	if term.IsTerminal(int(syscall.Stdin)) {
		fd = uintptr(syscall.Stdin)
	} else {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return "", fmt.Errorf("error allocating terminal: %w", err)
		}
		defer tty.Close()
		fd = tty.Fd()
	}
	_, err := out.WriteString(prompt)
	if err != nil {
		return "", err
	}
	pw, err := term.ReadPassword(int(fd))
	if err != nil {
		return "", err
	}
	_, err = out.WriteString("\n")
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

// Askpass prompts user for an interactive password.
// If confirm is true, confirms a second time.
// Mistakes during confirmation cause the process to restart upto a
// maximum of 2 times.
// We will try to use one of GUI askpass programs (in Linux/Unix)
// If none were found, will prompt in standard input
// If standard input is redirected, will open a new tty for reading password
func Askpass(prompt string, confirm bool, confirmPrompt string) (string, error) {
	var out *os.File
	if term.IsTerminal(int(syscall.Stdout)) {
		out = os.Stdout
	} else if term.IsTerminal(int(syscall.Stderr)) {
		out = os.Stderr
	} else {
		return "", fmt.Errorf("neither stdout not stderr is a terminal")
	}
	for i := 0; i < 2; i++ {
		pw1, err := askpassOnce(prompt, out)
		if err != nil {
			return "", err
		}
		if !confirm {
			return string(pw1), nil
		}

		pw2, err := askpassOnce(confirmPrompt, out)
		if err != nil {
			return "", err
		}

		a := string(pw1)
		b := string(pw2)
		if a == b {
			return a, nil
		}

		_, err = out.WriteString("** password mismatch; try again ..\n")
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("too many tries getting password")
}

func getAskpassBinaryPath() string {
	if os.Getenv("DISPLAY") == "" {
		return ""
	}
	if os.Getenv("GO_ASKPASS") != "" {
		askpassPath := os.Getenv("GO_ASKPASS")
		_, err := os.Stat(askpassPath)
		if err == nil {
			return askpassPath
		}
		log.Println(err)
	}
	for _, askpassPath := range []string{
		"/usr/lib/openssh/gnome-ssh-askpass",
		"/usr/bin/ksshaskpass",
		"/usr/bin/lxqt-openssh-askpass",
		"/usr/libexec/openssh/lxqt-openssh-askpass",
		"/usr/bin/ssh-askpass-fullscreen",
		"/usr/lib/ssh/x11-ssh-askpass",
	} {
		_, err := os.Stat(askpassPath)
		if err == nil {
			return askpassPath
		}
		if os.IsNotExist(err) {
			continue
		}
		log.Println(err)
	}
	return ""
}
