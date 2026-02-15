package commands

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// StatusBox renders a titled box with key-value fields.
//
//	StatusBox("Node", [][2]string{{"ID", "abc123"}, {"Port", "9735"}})
func StatusBox(title string, fields [][2]string) string {
	if !isTTY() {
		return statusBoxPlain(title, fields)
	}

	var sb strings.Builder
	sb.WriteString(StyleHeader.Render(title))
	sb.WriteString("\n")
	for _, f := range fields {
		label := StyleLabel.Render(f[0])
		value := StyleValue.Render(f[1])
		sb.WriteString(label + value + "\n")
	}

	return StyleBox.Render(strings.TrimRight(sb.String(), "\n"))
}

func statusBoxPlain(title string, fields [][2]string) string {
	var sb strings.Builder
	sb.WriteString(title + "\n")
	sb.WriteString(strings.Repeat("=", len(title)) + "\n")
	for _, f := range fields {
		sb.WriteString(fmt.Sprintf("%-14s %s\n", f[0]+":", f[1]))
	}
	return sb.String()
}

// RenderTable renders a styled table with headers and rows.
func RenderTable(headers []string, rows [][]string) string {
	if !isTTY() {
		return renderTablePlain(headers, rows)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorDim)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return StyleTableHeader
			}
			if row%2 == 0 {
				return StyleTableRow
			}
			return StyleTableRowAlt
		}).
		Headers(headers...).
		Rows(rows...)

	return t.String()
}

func renderTablePlain(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder
	// Header
	for i, h := range headers {
		sb.WriteString(fmt.Sprintf("%-*s  ", widths[i], h))
		_ = i
	}
	sb.WriteString("\n")
	// Separator
	for i, w := range widths {
		sb.WriteString(strings.Repeat("-", w))
		if i < len(widths)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")
	// Rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				sb.WriteString(fmt.Sprintf("%-*s  ", widths[i], cell))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Success prints a success message with a checkmark.
func Success(msg string) {
	if isTTY() {
		fmt.Println(StyleSuccess.Render("  " + msg))
	} else {
		fmt.Println("[OK] " + msg)
	}
}

// Error prints an error message with an X.
func Error(msg string) {
	if isTTY() {
		fmt.Println(StyleError.Render("  " + msg))
	} else {
		fmt.Println("[ERROR] " + msg)
	}
}

// Warning prints a warning message.
func Warning(msg string) {
	if isTTY() {
		fmt.Println(StyleWarning.Render("  " + msg))
	} else {
		fmt.Println("[WARN] " + msg)
	}
}

// Info prints an informational message.
func Info(msg string) {
	if isTTY() {
		fmt.Println(StyleInfo.Render("  " + msg))
	} else {
		fmt.Println("[INFO] " + msg)
	}
}

// WithSpinner runs a function while showing a spinner with the given message.
// Returns the error from the function.
func WithSpinner(msg string, fn func() error) error {
	if !isTTY() {
		fmt.Printf("%s...\n", msg)
		return fn()
	}

	var fnErr error
	err := spinner.New().
		Title(msg).
		Action(func() {
			fnErr = fn()
		}).
		Run()

	if err != nil {
		return err
	}
	return fnErr
}

// FormatBUNKER formats a token amount with thousands separators.
// amount is in base units (no decimals for BUNKER — it's an 18-decimal ERC-20,
// but for CLI display we show whole tokens).
func FormatBUNKER(amount *big.Int) string {
	if amount == nil {
		return "0 BUNKER"
	}

	// Convert from wei (18 decimals) to whole tokens
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	whole := new(big.Int).Div(amount, decimals)

	return addThousandsSep(whole.String()) + " BUNKER"
}

// FormatBUNKERWhole formats a whole token amount (no wei conversion).
func FormatBUNKERWhole(amount int64) string {
	s := fmt.Sprintf("%d", amount)
	return addThousandsSep(s) + " BUNKER"
}

func addThousandsSep(s string) string {
	if len(s) <= 3 {
		return s
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	// Insert commas from the right
	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	if negative {
		return "-" + result.String()
	}
	return result.String()
}

// FormatNodeID truncates a node ID for display.
func FormatNodeID(id string) string {
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "..." + id[len(id)-8:]
}

// FormatAddress truncates an Ethereum address for display.
func FormatAddress(addr string) string {
	if len(addr) <= 12 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

// SectionHeader renders a section header with a divider.
func SectionHeader(title string) string {
	if !isTTY() {
		return "\n" + title + "\n" + strings.Repeat("-", len(title))
	}
	return "\n" + StyleSubheader.Render(title)
}

// KeyValue renders a single key-value line with consistent alignment.
func KeyValue(key, value string) string {
	if !isTTY() {
		return fmt.Sprintf("  %-14s %s", key+":", value)
	}
	return "  " + StyleLabel.Render(key) + StyleValue.Render(value)
}

// Hint renders a dim hint/suggestion message.
func Hint(msg string) string {
	if !isTTY() {
		return "  " + msg
	}
	return "  " + StyleDim.Render(msg)
}

// DoctorCheck renders a doctor-style check result line.
func DoctorCheck(status string, name string, detail string) string {
	var icon string
	var style lipgloss.Style

	switch status {
	case "ok":
		icon = "  "
		style = StyleSuccess
	case "error":
		icon = "  "
		style = StyleError
	case "warning":
		icon = "  "
		style = StyleWarning
	case "skipped":
		icon = "  "
		style = StyleMuted
	default:
		icon = "  "
		style = StyleMuted
	}

	if !isTTY() {
		icons := map[string]string{"ok": "[OK]", "error": "[FAIL]", "warning": "[WARN]", "skipped": "[SKIP]"}
		i, ok := icons[status]
		if !ok {
			i = "[?]"
		}
		return fmt.Sprintf("  %-6s %-20s %s", i, name, detail)
	}

	nameStyled := lipgloss.NewStyle().Width(20).Render(name)
	return style.Render(icon) + nameStyled + StyleMuted.Render(detail)
}
