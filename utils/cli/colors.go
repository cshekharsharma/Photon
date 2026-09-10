// Package ansi allows for advanced terminal text manipulation.
package cli

import (
	"fmt"
	"io"
	"os"
)

type AnsiColor struct {
	bold  bool
	value int
}

var (
	Black      = AnsiColor{false, 30}
	BoldBlack  = AnsiColor{true, 30}
	Red        = AnsiColor{false, 31}
	BoldRed    = AnsiColor{true, 31}
	Green      = AnsiColor{false, 32}
	BoldGreen  = AnsiColor{true, 32}
	Yellow     = AnsiColor{false, 33}
	BoldYellow = AnsiColor{true, 33}
	Blue       = AnsiColor{false, 34}
	BoldBlue   = AnsiColor{true, 34}
	Purple     = AnsiColor{false, 35}
	BoldPurple = AnsiColor{true, 35}
	Cyan       = AnsiColor{false, 36}
	BoldCyan   = AnsiColor{true, 36}
	White      = AnsiColor{false, 37}
	BoldWhite  = AnsiColor{true, 37}
	Default    = AnsiColor{false, 39}

	Bold         = AnsiColor{false, 1}
	NotBold      = AnsiColor{false, 22}
	Underline    = AnsiColor{false, 4}
	NotUnderline = AnsiColor{false, 24}
	Reversed     = AnsiColor{false, 7}
	NotReversed  = AnsiColor{false, 27}
)

var (
	fgColor     = Default
	bgColor     = Default
	isBold      = false
	isUnderline = false
	isReversed  = false
)

// SetFg sets the foreground text color.
func SetFg(fg AnsiColor) {
	fgColor = fg
	if fg.bold {
		isBold = true
	}
	applySettings(os.Stdout)
}

// SetBg sets the background text color.
func SetBg(bg AnsiColor) {
	bgColor = bg
	applySettings(os.Stdout)
}

// SetBold sets the text to be bold or not.
func SetBold(bold bool) {
	isBold = bold
	applySettings(os.Stdout)
}

// SetUnderline sets the text to be underlined or not.
func SetUnderline(underline bool) {
	isUnderline = underline
	applySettings(os.Stdout)
}

// SetReversed sets the text to be reversed or not.
func SetReversed(reversed bool) {
	isReversed = reversed
	applySettings(os.Stdout)
}

// SetAll sets the foreground, background, bold, underline and reversed properties at once.
func SetAll(fg, bg AnsiColor, bold, underline, reversed bool) {
	fgColor = fg
	bgColor = bg
	isBold = bold
	isUnderline = underline
	isReversed = reversed
	applySettings(os.Stdout)
}

// ResetAll resets all properties to default.
func ResetAll() {
	SetAll(Default, Default, false, false, false)
	resetLine(os.Stdout)
}

func resetLine(output io.Writer) {
	if _, err := fmt.Fprint(output, "\033[K"); err != nil {
		return
	}
}

func applySettings(output io.Writer) {
	b := 22
	if isBold {
		b = 1
	}
	u := 24
	if isUnderline {
		u = 4
	}
	r := 27
	if isReversed {
		r = 7
	}
	if _, err := fmt.Fprintf(output, "\033[0;%d;%d;%d;%d;%dm", fgColor.value, bgColor.value+10, b, u, r); err != nil {
		return
	}
}

// Print is a replacement for fmt.Print, which accepts an AnsiColor as the first argument.
func Print(col AnsiColor, s ...any) {
	Fprint(os.Stdout, col, s...)
}

// Println is a replacement for fmt.Println, which accepts an AnsiColor as the first argument.
func Println(col AnsiColor, s ...any) {
	Fprintln(os.Stdout, col, s...)
}

// Printf is a replacement for fmt.Printf, which accepts an AnsiColor as the first argument.
func Printf(col AnsiColor, s string, args ...any) {
	Fprintf(os.Stdout, col, s, args...)
}

// Fprint is a replacement for fmt.Fprint, which accepts an AnsiColor as the first argument.
func Fprint(output io.Writer, col AnsiColor, s ...any) {
	writeWithColor(output, col, func(output io.Writer) (int, error) {
		return fmt.Fprint(output, s...)
	})
}

// Fprintln is a replacement for fmt.Fprintln, which accepts an AnsiColor as the first argument.
func Fprintln(output io.Writer, col AnsiColor, s ...any) {
	writeWithColor(output, col, func(output io.Writer) (int, error) {
		return fmt.Fprintln(output, s...)
	})
}

// Fprintf is a replacement for fmt.Fprintf, which accepts an AnsiColor as the first argument.
func Fprintf(output io.Writer, col AnsiColor, s string, args ...any) {
	writeWithColor(output, col, func(output io.Writer) (int, error) {
		return fmt.Fprintf(output, s, args...)
	})
}

func writeWithColor(output io.Writer, col AnsiColor, write func(io.Writer) (int, error)) {
	if _, err := fmt.Fprintf(output, "\033[0;%dm", col.value); err != nil {
		return
	}
	if col.bold {
		if _, err := fmt.Fprintf(output, "\033[1m"); err != nil {
			return
		}
	}
	if _, err := write(output); err != nil {
		return
	}
	applySettings(output)
}

// Add these functions to cli.go
func GetFgColor() AnsiColor {
	return fgColor
}

func GetBgColor() AnsiColor {
	return bgColor
}

func IsBold() bool {
	return isBold
}

func IsUnderline() bool {
	return isUnderline
}

func IsReversed() bool {
	return isReversed
}
