package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var monitorRefresh int

func NewMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Full-screen monitoring dashboard",
		Long: `Interactive monitoring dashboard with live updates.

Tabs: [Overview] [Peers] [Containers]
Keys: tab=switch views  q=quit  r=refresh`,
		RunE: runMonitor,
	}

	cmd.Flags().IntVarP(&monitorRefresh, "refresh", "r", 2, "Refresh interval in seconds")

	return cmd
}

func runMonitor(cmd *cobra.Command, args []string) error {
	dc := client.NewDaemonClient(SocketPath)
	if err := dc.Connect(); err != nil {
		return fmt.Errorf("daemon not running. Start with: moltbunker start")
	}

	p := tea.NewProgram(
		newMonitorModel(dc, time.Duration(monitorRefresh)*time.Second),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// --- bubbletea model ---

type tab int

const (
	tabOverview    tab = 0
	tabPeers       tab = 1
	tabContainers  tab = 2
)

var tabNames = []string{"Overview", "Peers", "Containers"}

type monitorModel struct {
	client   *client.DaemonClient
	interval time.Duration

	activeTab  tab
	status     *client.StatusResponse
	peers      []client.PeerInfo
	containers []client.ContainerInfo
	lastUpdate time.Time
	err        error
	width      int
	height     int
}

type tickMsg time.Time
type dataMsg struct {
	status     *client.StatusResponse
	peers      []client.PeerInfo
	containers []client.ContainerInfo
	err        error
}

func newMonitorModel(dc *client.DaemonClient, interval time.Duration) monitorModel {
	return monitorModel{
		client:   dc,
		interval: interval,
	}
}

func (m monitorModel) Init() tea.Cmd {
	return tea.Batch(fetchData(m.client), tick(m.interval))
}

func (m monitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % tab(len(tabNames))
		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab - 1 + tab(len(tabNames))) % tab(len(tabNames))
		case "r":
			return m, fetchData(m.client)
		}

	case tickMsg:
		return m, tea.Batch(fetchData(m.client), tick(m.interval))

	case dataMsg:
		m.status = msg.status
		m.peers = msg.peers
		m.containers = msg.containers
		m.err = msg.err
		m.lastUpdate = time.Now()
	}

	return m, nil
}

func (m monitorModel) View() string {
	var sb strings.Builder

	// Title
	title := StyleAccent.Bold(true).Render(" MOLTBUNKER MONITOR ")
	sb.WriteString(title)
	sb.WriteString("\n")

	// Tab bar
	sb.WriteString(renderTabs(m.activeTab))
	sb.WriteString("\n\n")

	// Content
	if m.err != nil {
		sb.WriteString(StyleError.Render(fmt.Sprintf("  Error: %v", m.err)))
	} else if m.status == nil {
		sb.WriteString(StyleMuted.Render("  Loading..."))
	} else {
		switch m.activeTab {
		case tabOverview:
			sb.WriteString(m.viewOverview())
		case tabPeers:
			sb.WriteString(m.viewPeers())
		case tabContainers:
			sb.WriteString(m.viewContainers())
		}
	}

	// Footer
	sb.WriteString("\n\n")
	footer := StyleDim.Render(fmt.Sprintf(
		"  tab=switch  q=quit  r=refresh  |  Updated: %s",
		m.lastUpdate.Format("15:04:05"),
	))
	sb.WriteString(footer)

	return sb.String()
}

func renderTabs(active tab) string {
	var tabs []string
	for i, name := range tabNames {
		if tab(i) == active {
			tabs = append(tabs, lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(ColorAccent).
				Padding(0, 2).
				Render(name))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2).
				Render(name))
		}
	}
	return "  " + strings.Join(tabs, " ")
}

func (m monitorModel) viewOverview() string {
	s := m.status
	var sb strings.Builder

	statusStr := "STOPPED"
	if s.Running {
		statusStr = "RUNNING"
	}

	sb.WriteString(StatusBox("Node", [][2]string{
		{"Status", statusStr},
		{"Node ID", FormatNodeID(s.NodeID)},
		{"Version", s.Version},
		{"Port", fmt.Sprintf("%d", s.Port)},
		{"Region", s.Region},
	}))
	sb.WriteString("\n")

	sb.WriteString(StatusBox("Network", [][2]string{
		{"Nodes", fmt.Sprintf("%d", s.NetworkNodes)},
		{"Peers", fmt.Sprintf("%d", len(m.peers))},
		{"Containers", fmt.Sprintf("%d", s.Containers)},
	}))
	sb.WriteString("\n")

	torStatus := "Disabled"
	if s.TorEnabled {
		torStatus = "Enabled"
	}
	torFields := [][2]string{{"Status", torStatus}}
	if s.TorAddress != "" {
		torFields = append(torFields, [2]string{"Onion", s.TorAddress})
	}
	sb.WriteString(StatusBox("Tor", torFields))

	return sb.String()
}

func (m monitorModel) viewPeers() string {
	if len(m.peers) == 0 {
		return StyleMuted.Render("  No connected peers.")
	}

	headers := []string{"Peer ID", "Region", "Address", "Last Seen"}
	var rows [][]string
	for _, p := range m.peers {
		region := p.Region
		if region == "" {
			region = "unknown"
		}
		rows = append(rows, []string{
			FormatNodeID(p.ID),
			region,
			p.Address,
			p.LastSeen.Format("15:04:05"),
		})
	}

	result := RenderTable(headers, rows)
	result += "\n" + StyleMuted.Render(fmt.Sprintf("  %d peer(s)", len(m.peers)))
	return result
}

func (m monitorModel) viewContainers() string {
	if len(m.containers) == 0 {
		return StyleMuted.Render("  No containers running.")
	}

	headers := []string{"ID", "Image", "Status", "Regions", "Created"}
	var rows [][]string
	for _, ct := range m.containers {
		regions := "-"
		if len(ct.Regions) > 0 {
			regions = ct.Regions[0]
			if len(ct.Regions) > 1 {
				regions += fmt.Sprintf(" +%d", len(ct.Regions)-1)
			}
		}
		rows = append(rows, []string{
			FormatNodeID(ct.ID),
			ct.Image,
			ct.Status,
			regions,
			ct.CreatedAt.Format("Jan 02 15:04"),
		})
	}

	result := RenderTable(headers, rows)
	result += "\n" + StyleMuted.Render(fmt.Sprintf("  %d container(s)", len(m.containers)))
	return result
}

// --- commands ---

func fetchData(dc *client.DaemonClient) tea.Cmd {
	return func() tea.Msg {
		var d dataMsg

		status, err := dc.Status()
		if err != nil {
			d.err = err
			return d
		}
		d.status = status

		peers, _ := dc.GetPeers()
		d.peers = peers

		containers, _ := dc.List()
		d.containers = containers

		return d
	}
}

func tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
