// Package termcolor provides terminal color formatting helpers.
package termcolor

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// ANSI escape codes.
const (
	reset = "\033[0m"
	bold  = "\033[1m"
	dim   = "\033[2m"
	red   = "\033[91m"
	green = "\033[32m"
)

var enabled = term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("NO_COLOR") == ""

// Disable turns off color output (for --no-color flag).
func Disable() {
	enabled = false
}

// Enabled returns whether color output is enabled.
func Enabled() bool {
	return enabled
}

func wrap(s, code string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

// Bold returns s wrapped in bold.
func Bold(s string) string {
	return wrap(s, bold)
}

// Dim returns s wrapped in dim.
func Dim(s string) string {
	return wrap(s, dim)
}

// Red returns s wrapped in red.
func Red(s string) string {
	return wrap(s, red)
}

// Green returns s wrapped in green.
func Green(s string) string {
	return wrap(s, green)
}

// Header returns a centered, padded header like "===== text =====".
func Header(text, char string) string {
	return colorHeader(text, char, bold)
}

// RedHeader returns a centered, padded header in bold red.
func RedHeader(text, char string) string {
	return colorHeader(text, char, bold+red)
}

// GreenHeader returns a centered, padded header in bold green.
func GreenHeader(text, char string) string {
	return colorHeader(text, char, bold+green)
}

func colorHeader(text, char, code string) string {
	if char == "" {
		char = "-"
	}
	textWithSpaces := " " + text + " "
	width := termWidth()
	// Ensure some sane minimal width to keep visual separation
	if width < len(textWithSpaces)+2 {
		return wrap(char+textWithSpaces+char, code)
	}
	totalPadding := width - len(textWithSpaces)
	leftPad := totalPadding / 2
	rightPad := totalPadding - leftPad
	line := strings.Repeat(char, leftPad) + textWithSpaces + strings.Repeat(char, rightPad)
	return wrap(line, code)
}

// termWidth returns the current terminal width in columns.
// Prefers TTY size, falls back to $COLUMNS, then 80.
func termWidth() int {
	fd := int(os.Stdout.Fd()) //nolint:gosec // G115: file descriptors are always small.
	if term.IsTerminal(fd) {
		if w, _, err := term.GetSize(fd); err == nil && w > 0 {
			return w
		}
	}
	if v := os.Getenv("COLUMNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 80
}
