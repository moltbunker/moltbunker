package doctor

import (
	"bytes"
	"context"
	"testing"
)

func TestDoctorReport(t *testing.T) {
	ctx := context.Background()

	opts := DoctorOptions{
		JSON: false,
	}

	var buf bytes.Buffer
	d := NewWithWriter(opts, &buf, false)

	// Run the doctor checks
	report, err := d.Run(ctx)
	if err != nil {
		t.Fatalf("Doctor.Run() failed: %v", err)
	}

	// Verify we got results
	if len(report.Checks) == 0 {
		t.Error("Expected at least one check result")
	}

	// Verify summary is populated
	if report.Summary.Total != len(report.Checks) {
		t.Errorf("Summary.Total = %d, want %d", report.Summary.Total, len(report.Checks))
	}

	// Verify passed + failed + warned + skipped = total
	sum := report.Summary.Passed + report.Summary.Failed + report.Summary.Warned + report.Summary.Skipped
	if sum != report.Summary.Total {
		t.Errorf("Sum of results (%d) != Total (%d)", sum, report.Summary.Total)
	}
}

func TestDoctorWithCategoryFilter(t *testing.T) {
	ctx := context.Background()

	opts := DoctorOptions{
		Category: CategoryRuntime,
	}

	var buf bytes.Buffer
	d := NewWithWriter(opts, &buf, false)

	report, err := d.Run(ctx)
	if err != nil {
		t.Fatalf("Doctor.Run() failed: %v", err)
	}

	// All results should be in the runtime category
	for _, check := range report.Checks {
		if check.Category != CategoryRuntime {
			t.Errorf("Expected category %s, got %s for check %s", CategoryRuntime, check.Category, check.Name)
		}
	}
}

func TestDoctorJSONOutput(t *testing.T) {
	ctx := context.Background()

	opts := DoctorOptions{
		JSON: true,
	}

	d := New(opts)

	report, err := d.Run(ctx)
	if err != nil {
		t.Fatalf("Doctor.Run() failed: %v", err)
	}

	if report == nil {
		t.Error("Expected non-nil report")
	}
}

func TestSummaryIsHealthy(t *testing.T) {
	tests := []struct {
		name    string
		summary Summary
		want    bool
	}{
		{
			name:    "all passed",
			summary: Summary{Total: 5, Passed: 5, Failed: 0},
			want:    true,
		},
		{
			name:    "has failures",
			summary: Summary{Total: 5, Passed: 3, Failed: 2},
			want:    false,
		},
		{
			name:    "only warnings",
			summary: Summary{Total: 5, Passed: 3, Warned: 2},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.summary.IsHealthy(); got != tt.want {
				t.Errorf("Summary.IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckResult(t *testing.T) {
	result := CheckResult{
		Name:       "Test Check",
		Category:   CategoryRuntime,
		Status:     StatusOK,
		Message:    "All good",
		Fixable:    false,
	}

	if result.Name != "Test Check" {
		t.Errorf("Expected name 'Test Check', got '%s'", result.Name)
	}

	if result.Status != StatusOK {
		t.Errorf("Expected status OK, got %s", result.Status)
	}
}

func TestOutput(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput(&buf, false)

	out.Header()
	if !bytes.Contains(buf.Bytes(), []byte("Moltbunker Doctor")) {
		t.Error("Header should contain 'Moltbunker Doctor'")
	}

	buf.Reset()
	out.CheckStart(1, 8, "Test check")
	if !bytes.Contains(buf.Bytes(), []byte("[1/8]")) {
		t.Error("CheckStart should contain progress indicator")
	}

	buf.Reset()
	result := CheckResult{
		Status:  StatusOK,
		Message: "Test passed",
	}
	out.CheckResult(result)
	if !bytes.Contains(buf.Bytes(), []byte("✓")) {
		t.Error("CheckResult with StatusOK should contain checkmark")
	}

	buf.Reset()
	result = CheckResult{
		Status:  StatusError,
		Message: "Test failed",
	}
	out.CheckResult(result)
	if !bytes.Contains(buf.Bytes(), []byte("✗")) {
		t.Error("CheckResult with StatusError should contain X mark")
	}
}
