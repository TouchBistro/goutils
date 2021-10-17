// Package color provides functions for creating colored strings.
//
// There are several functions provided to make it easy to set foreground colors.
//
// 	// creates a string with a red foreground color
// 	color.Red("uh oh")
//
// Colors can be globally enabled or disabled by using SetEnabled.
//
// This package also supports the NO_COLOR environment variable.
// If NO_COLOR is set with any value, colors will be disabled.
// See https://no-color.org for more details.
package color

import (
	"fmt"
	"os"
	"regexp"
)

type ansiCode uint8

const (
	fgBlack ansiCode = iota + 30
	fgRed
	fgGreen
	fgYellow
	fgBlue
	fgMagenta
	fgCyan
	fgWhite
	_ // skip value
	fgReset
)

// Support for NO_COLOR env var
// https://no-color.org/
var (
	noColor = false
	enabled bool
)

func init() {
	// The standard says the value doesn't matter, only whether or not it's set
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		noColor = true
	}
	enabled = !noColor
}

func apply(s string, start, end ansiCode) string {
	if !enabled {
		return s
	}

	regex := regexp.MustCompile(fmt.Sprintf("\\x1b\\[%dm", end))
	// Remove any occurrences of reset to make sure color isn't messed up
	sanitized := regex.ReplaceAllString(s, "")
	return fmt.Sprintf("\x1b[%dm%s\x1b[%dm", start, sanitized, end)
}

// SetEnabled sets whether color is enabled or disabled.
// If the NO_COLOR environment variable is set, this function will
// do nothing as NO_COLOR takes precedence.
func SetEnabled(e bool) {
	// NO_COLOR overrides this
	if noColor {
		return
	}
	enabled = e
}

// Black creates a black colored string.
func Black(s string) string {
	return apply(s, fgBlack, fgReset)
}

// Red creates a red colored string.
func Red(s string) string {
	return apply(s, fgRed, fgReset)
}

// Green creates a green colored string.
func Green(s string) string {
	return apply(s, fgGreen, fgReset)
}

// Yellow creates a yellow colored string.
func Yellow(s string) string {
	return apply(s, fgYellow, fgReset)
}

// Blue creates a blue colored string.
func Blue(s string) string {
	return apply(s, fgBlue, fgReset)
}

// Magenta creates a magenta colored string.
func Magenta(s string) string {
	return apply(s, fgMagenta, fgReset)
}

// Cyan creates a cyan colored string.
func Cyan(s string) string {
	return apply(s, fgCyan, fgReset)
}

// White creates a white colored string.
func White(s string) string {
	return apply(s, fgWhite, fgReset)
}
