package doctor

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// Output handles formatted output for the doctor command
type Output struct {
	writer    io.Writer
	useColors bool
}

// NewOutput creates a new Output instance
func NewOutput(w io.Writer, useColors bool) *Output {
	if w == nil {
		w = os.Stdout
	}
	return &Output{
		writer:    w,
		useColors: useColors,
	}
}

// Header prints the doctor header
func (o *Output) Header() {
	o.println("")
	o.printlnBold("Moltbunker Doctor")
	o.println(strings.Repeat("=", 17))
	o.println("")
}

// CheckStart prints the start of a check
func (o *Output) CheckStart(index, total int, name string) {
	o.printf("[%d/%d] Checking %s...\n", index, total, name)
}

// CheckResult prints the result of a check
func (o *Output) CheckResult(result CheckResult) {
	var icon, color string
	switch result.Status {
	case StatusOK:
		icon = "✓"
		color = colorGreen
	case StatusWarning:
		icon = "!"
		color = colorYellow
	case StatusError:
		icon = "✗"
		color = colorRed
	case StatusSkipped:
		icon = "-"
		color = colorDim
	}

	if o.useColors {
		o.printf("  %s%s%s %s\n", color, icon, colorReset, result.Message)
	} else {
		o.printf("  %s %s\n", icon, result.Message)
	}

	if result.Details != "" {
		o.printf("    %s\n", result.Details)
	}

	if result.Status == StatusError && result.Fixable && result.FixCommand != "" {
		o.printf("    Fix: %s\n", result.FixCommand)
	}
}

// Summary prints the summary at the end
func (o *Output) Summary(summary Summary) {
	o.println("")
	if o.useColors {
		o.printf("Summary: %s%d passed%s, ", colorGreen, summary.Passed, colorReset)
		if summary.Failed > 0 {
			o.printf("%s%d failed%s", colorRed, summary.Failed, colorReset)
		} else {
			o.printf("0 failed")
		}
		if summary.Warned > 0 {
			o.printf(", %s%d warnings%s", colorYellow, summary.Warned, colorReset)
		}
		if summary.Fixable > 0 {
			o.printf(", %s%d fixable%s", colorCyan, summary.Fixable, colorReset)
		}
		o.println("")
	} else {
		o.printf("Summary: %d passed, %d failed", summary.Passed, summary.Failed)
		if summary.Warned > 0 {
			o.printf(", %d warnings", summary.Warned)
		}
		if summary.Fixable > 0 {
			o.printf(", %d fixable", summary.Fixable)
		}
		o.println("")
	}

	if summary.Fixable > 0 && summary.Failed > 0 {
		o.println("Run 'moltbunker doctor --fix' to install missing dependencies.")
	}
}

// DryRunInfo prints info about dry run mode
func (o *Output) DryRunInfo() {
	if o.useColors {
		o.printf("%s[DRY RUN]%s The following actions would be performed:\n\n", colorCyan, colorReset)
	} else {
		o.println("[DRY RUN] The following actions would be performed:")
		o.println("")
	}
}

// FixAction prints an action that would be or is being performed
func (o *Output) FixAction(pkg string, dryRun bool) {
	if dryRun {
		o.printf("  Would install: %s\n", pkg)
	} else {
		o.printf("  Installing: %s\n", pkg)
	}
}

// FixSuccess prints a successful fix
func (o *Output) FixSuccess(pkg string) {
	if o.useColors {
		o.printf("  %s✓%s Installed: %s\n", colorGreen, colorReset, pkg)
	} else {
		o.printf("  ✓ Installed: %s\n", pkg)
	}
}

// FixError prints a fix error
func (o *Output) FixError(pkg string, err error) {
	if o.useColors {
		o.printf("  %s✗%s Failed to install %s: %v\n", colorRed, colorReset, pkg, err)
	} else {
		o.printf("  ✗ Failed to install %s: %v\n", pkg, err)
	}
}

// Error prints an error message
func (o *Output) Error(msg string) {
	if o.useColors {
		o.printf("%sError:%s %s\n", colorRed, colorReset, msg)
	} else {
		o.printf("Error: %s\n", msg)
	}
}

// Info prints an info message
func (o *Output) Info(msg string) {
	if o.useColors {
		o.printf("%s%s%s\n", colorDim, msg, colorReset)
	} else {
		o.println(msg)
	}
}

func (o *Output) println(s string) {
	fmt.Fprintln(o.writer, s)
}

func (o *Output) printlnBold(s string) {
	if o.useColors {
		fmt.Fprintf(o.writer, "%s%s%s\n", colorBold, s, colorReset)
	} else {
		fmt.Fprintln(o.writer, s)
	}
}

func (o *Output) printf(format string, args ...interface{}) {
	fmt.Fprintf(o.writer, format, args...)
}
