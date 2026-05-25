// Package cli provides colored, formatted CLI output for user-facing messages.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

var (
	// Output writer (can be changed for testing)
	Out io.Writer = os.Stdout
	Err io.Writer = os.Stderr

	// Colors
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	infoColor    = color.New(color.FgBlue)
	subtleColor  = color.New(color.FgHiBlack)
	headerColor  = color.New(color.FgBlue, color.Bold)
	labelColor   = color.New(color.FgCyan)
	pathColor    = color.New(color.FgWhite)
)

// Output represents the CLI output handler
type Output struct {
	w       io.Writer
	verbose bool
	noColor bool
}

// New creates a new Output instance
func New() *Output {
	return &Output{
		w:       Out,
		verbose: os.Getenv("VERBOSE") != "" || os.Getenv("DEBUG") != "",
		noColor: os.Getenv("NO_COLOR") != "",
	}
}

// SetWriter sets the output writer
func (o *Output) SetWriter(w io.Writer) {
	o.w = w
}

func (o *Output) Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	f.BoolVarP(&o.verbose, "verbose", "v", o.verbose, "enable verbose output")
	f.BoolVarP(&o.noColor, "no-color", "n", o.noColor, "disable color output")

	return f
}

// Configure executes side-effects structurally based on the natively parsed state.
func (o *Output) Configure() {
	if o.noColor {
		color.NoColor = true
	}
}

// Writer returns the underlying output writer.
func (o *Output) Writer() io.Writer {
	return o.w
}

// IsVerbose returns whether verbose mode is enabled
func (o *Output) IsVerbose() bool {
	return o.verbose
}

// Success prints a success message
func (o *Output) Success(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = successColor.Fprintln(o.w, o.prefix("✓")+" "+msg)
}

// Error prints an error message
func (o *Output) Error(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = errorColor.Fprintln(o.w, o.prefix("✗")+" "+msg)
}

// Warn prints a warning message
func (o *Output) Warn(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = warnColor.Fprintln(o.w, o.prefix("!")+" "+msg)
}

// Info prints an info message
func (o *Output) Info(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = infoColor.Fprintln(o.w, o.prefix("•")+" "+msg)
}

// Header prints a header message
func (o *Output) Header(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = headerColor.Fprintln(o.w, "\n"+msg)
}

// Subtle prints a subtle/secondary message
func (o *Output) Subtle(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = subtleColor.Fprintln(o.w, "  "+msg)
}

// Label prints a labeled value
func (o *Output) Label(label, value string) {
	_, _ = labelColor.Fprintf(o.w, "  %-15s ", label+":")
	_, _ = pathColor.Fprintln(o.w, value)
}

// Verbose prints a message only if verbose mode is enabled
func (o *Output) Verbose(format string, args ...any) {
	if !o.verbose {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = subtleColor.Fprintln(o.w, "  "+msg)
}

// Block prints an indented block of text
func (o *Output) Block(text string) {
	lines := strings.SplitSeq(strings.TrimSpace(text), "\n")
	for line := range lines {
		_, _ = subtleColor.Fprintln(o.w, "  "+line)
	}
}

// Raw prints raw text without formatting
func (o *Output) Raw(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = fmt.Fprintln(o.w, msg)
}

// Fatal prints an error and exits
func (o *Output) Fatal(format string, args ...any) {
	o.Error(format, args...)
	os.Exit(1)
}

// prefix returns a colored prefix symbol
func (o *Output) prefix(symbol string) string {
	if o.noColor {
		return symbol
	}

	switch symbol {
	case "✓":
		return successColor.Sprint(symbol)
	case "✗":
		return errorColor.Sprint(symbol)
	case "!":
		return warnColor.Sprint(symbol)
	case "•":
		return infoColor.Sprint(symbol)
	default:
		return symbol
	}
}

// Package-level convenience functions (use default output)

var defaultOutput = New()

// Default returns the default output handler
func Default() *Output {
	return defaultOutput
}

// Success prints a success message using the default output
func Success(format string, args ...any) {
	defaultOutput.Success(format, args...)
}

// Error prints an error message using the default output
func Error(format string, args ...any) {
	defaultOutput.Error(format, args...)
}

// Warn prints a warning message using the default output
func Warn(format string, args ...any) {
	defaultOutput.Warn(format, args...)
}

// Info prints an info message using the default output
func Info(format string, args ...any) {
	defaultOutput.Info(format, args...)
}

// Header prints a header message using the default output
func Header(format string, args ...any) {
	defaultOutput.Header(format, args...)
}

// Subtle prints a subtle message using the default output
func Subtle(format string, args ...any) {
	defaultOutput.Subtle(format, args...)
}

// Label prints a labeled value using the default output
func Label(label, value string) {
	defaultOutput.Label(label, value)
}

// Verbose prints a verbose message using the default output
func Verbose(format string, args ...any) {
	defaultOutput.Verbose(format, args...)
}

// Block prints a block of text using the default output
func Block(text string) {
	defaultOutput.Block(text)
}

// Raw prints raw text using the default output
func Raw(format string, args ...any) {
	defaultOutput.Raw(format, args...)
}

// Fatal prints an error and exits using the default output
func Fatal(format string, args ...any) {
	defaultOutput.Fatal(format, args...)
}
