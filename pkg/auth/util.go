package auth

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)

// ReadString reads a string from the stdin.
func ReadString(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}
	r := bufio.NewReader(os.Stdin)
	s, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	s = strings.TrimRight(s, "\r\n")
	return s, nil
}

// ReadPassword reads a password from the stdin without echo back.
func ReadPassword(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}
	b, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	if _, err := fmt.Fprintln(os.Stderr); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}
	return string(b), nil
}
