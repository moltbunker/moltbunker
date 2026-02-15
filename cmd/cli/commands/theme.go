package commands

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Brand colors
var (
	ColorAccent  = lipgloss.Color("#ef4444") // Red accent
	ColorSuccess = lipgloss.Color("#22c55e") // Green
	ColorWarning = lipgloss.Color("#eab308") // Yellow
	ColorError   = lipgloss.Color("#ef4444") // Red
	ColorInfo    = lipgloss.Color("#3b82f6") // Blue
	ColorMuted   = lipgloss.Color("#6b7280") // Gray
	ColorDim     = lipgloss.Color("#4b5563") // Darker gray
	ColorWhite   = lipgloss.Color("#f9fafb") // Off-white
)

// isTTY reports whether stdout is a terminal.
func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Semantic text styles
var (
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	StyleSubheader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMuted)

	StyleAccent = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorError)

	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorInfo)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleBold = lipgloss.NewStyle().
			Bold(true)

	StyleLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(14)

	StyleValue = lipgloss.NewStyle().
			Foreground(ColorWhite)
)

// Box styles
var (
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim).
			Padding(0, 1)

	StyleBoxAccent = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(0, 1)

	StyleBoxSuccess = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1)

	StyleBoxError = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 1)
)

// Table header style
var (
	StyleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent).
				Padding(0, 1)

	StyleTableRow = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Padding(0, 1)

	StyleTableRowAlt = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1)
)

// Status badge styles
func StatusBadge(status string) string {
	switch status {
	case "running", "active", "healthy", "ok":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorSuccess).
			Padding(0, 1).
			Bold(true).
			Render(status)
	case "stopped", "inactive", "error", "failed":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorError).
			Padding(0, 1).
			Bold(true).
			Render(status)
	case "pending", "deploying", "warning":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorWarning).
			Padding(0, 1).
			Bold(true).
			Render(status)
	default:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorMuted).
			Padding(0, 1).
			Render(status)
	}
}

// Logo returns the styled moltbunker logo/brand text
func Logo() string {
	return StyleAccent.Render("moltbunker")
}
