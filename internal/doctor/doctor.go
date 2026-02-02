package doctor

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"runtime"
)

// Doctor orchestrates health checks for the moltbunker system
type Doctor struct {
	checkers       []Checker
	packageManager PackageManager
	output         *Output
	options        DoctorOptions
}

// New creates a new Doctor instance with default checkers for the current platform
func New(opts DoctorOptions) *Doctor {
	d := &Doctor{
		options: opts,
	}

	// Determine if we should use colors
	useColors := !opts.JSON && isTerminal(os.Stdout)

	d.output = NewOutput(os.Stdout, useColors)

	// Initialize package manager based on platform
	if runtime.GOOS == "darwin" {
		d.packageManager = NewHomebrewManager()
	}

	// Register default checkers based on platform
	d.registerDefaultCheckers()

	return d
}

// NewWithWriter creates a Doctor with a custom writer (useful for testing)
func NewWithWriter(opts DoctorOptions, w io.Writer, useColors bool) *Doctor {
	d := &Doctor{
		options: opts,
		output:  NewOutput(w, useColors),
	}

	if runtime.GOOS == "darwin" {
		d.packageManager = NewHomebrewManager()
	}

	d.registerDefaultCheckers()

	return d
}

// registerDefaultCheckers adds all platform-specific checkers
func (d *Doctor) registerDefaultCheckers() {
	if runtime.GOOS == "darwin" {
		d.registerDarwinCheckers()
	}
}

// registerDarwinCheckers adds macOS-specific checkers
func (d *Doctor) registerDarwinCheckers() {
	d.checkers = []Checker{
		// Runtime checks
		NewGoVersionChecker(),
		NewColimaChecker(), // Check Colima before containerd on macOS
		NewContainerdChecker(),

		// Services checks
		NewTorChecker(),
		NewIPFSChecker(),

		// System checks
		NewDiskSpaceChecker(),
		NewMemoryChecker(),

		// Config checks
		NewConfigFileChecker(),

		// Permission checks
		NewSocketPermissionChecker(),
	}
}

// AddChecker adds a custom checker
func (d *Doctor) AddChecker(c Checker) {
	d.checkers = append(d.checkers, c)
}

// Run executes all checks and returns a report
func (d *Doctor) Run(ctx context.Context) (*DoctorReport, error) {
	report := &DoctorReport{
		Checks: make([]CheckResult, 0, len(d.checkers)),
	}

	// Filter checkers by category if specified
	checkers := d.filterCheckers()

	if d.options.JSON {
		// JSON mode - run silently and output JSON at the end
		for _, checker := range checkers {
			result := checker.Check(ctx)
			report.Checks = append(report.Checks, result)
			d.updateSummary(&report.Summary, result)
		}
		return report, d.outputJSON(report)
	}

	// Interactive mode
	d.output.Header()

	if d.options.DryRun && d.options.Fix {
		d.output.DryRunInfo()
	}

	for i, checker := range checkers {
		d.output.CheckStart(i+1, len(checkers), checker.Name())
		result := checker.Check(ctx)
		d.output.CheckResult(result)
		report.Checks = append(report.Checks, result)
		d.updateSummary(&report.Summary, result)

		// Handle auto-fix if enabled (for both errors and fixable warnings)
		shouldFix := d.options.Fix && checker.CanFix() && result.Fixable &&
			(result.Status == StatusError || result.Status == StatusWarning)
		if shouldFix {
			if d.options.DryRun {
				d.output.FixAction(result.FixPackage, true)
			} else if d.packageManager != nil {
				d.output.FixAction(result.FixPackage, false)
				if err := checker.Fix(ctx, d.packageManager); err != nil {
					d.output.FixError(result.FixPackage, err)
				} else {
					d.output.FixSuccess(result.FixPackage)
					// Recheck after fix
					newResult := checker.Check(ctx)
					if newResult.Status == StatusOK {
						// Update the result and summary
						report.Checks[len(report.Checks)-1] = newResult
						if result.Status == StatusError {
							report.Summary.Failed--
						} else {
							report.Summary.Warned--
						}
						report.Summary.Passed++
						report.Summary.Fixable--
					}
				}
			}
		}
	}

	d.output.Summary(report.Summary)

	return report, nil
}

// filterCheckers returns checkers filtered by category if specified
func (d *Doctor) filterCheckers() []Checker {
	if d.options.Category == "" {
		return d.checkers
	}

	filtered := make([]Checker, 0)
	for _, c := range d.checkers {
		if c.Category() == d.options.Category {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// updateSummary updates the summary based on a check result
func (d *Doctor) updateSummary(summary *Summary, result CheckResult) {
	summary.Total++
	switch result.Status {
	case StatusOK:
		summary.Passed++
	case StatusError:
		summary.Failed++
		if result.Fixable {
			summary.Fixable++
		}
	case StatusWarning:
		summary.Warned++
	case StatusSkipped:
		summary.Skipped++
	}
}

// outputJSON outputs the report as JSON
func (d *Doctor) outputJSON(report *DoctorReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		fi, err := f.Stat()
		if err != nil {
			return false
		}
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}
