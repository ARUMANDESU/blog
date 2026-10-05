package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	prog := filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Generate a bcrypt hash for the admin password.

Usage:
  %[1]s [password]
  echo 'password' | %[1]s
  %[1]s                      (interactive prompt, no echo)

Options:
  -h, --help    show this help'
`, prog)
	}
	flag.Parse() // -h / --help prints usage and exits 0

	pw, err := readPassword(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
		flag.Usage()
		os.Exit(2)
	}

	hash, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to generate hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(hash))
}

func readPassword(args []string) ([]byte, error) {
	switch len(args) {
	case 0:
		// fall through to stdin
	case 1:
		if args[0] == "" {
			return nil, fmt.Errorf("password is empty")
		}
		return []byte(args[0]), nil
	default:
		return nil, fmt.Errorf("too many arguments (quote passwords containing spaces)")
	}

	fd := int(os.Stdin.Fd())

	if term.IsTerminal(fd) {
		fmt.Fprint(os.Stderr, "Password: ")
		pw, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return nil, fmt.Errorf("reading password: %w", err)
		}
		if len(pw) == 0 {
			return nil, fmt.Errorf("password is empty")
		}
		return pw, nil
	}

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return nil, fmt.Errorf("no password on stdin")
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return nil, fmt.Errorf("password is empty")
	}
	return []byte(line), nil
}
