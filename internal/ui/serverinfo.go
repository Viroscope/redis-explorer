package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"redis-explorer/internal/redis"
)

// ServerInfo represents the server info panel
type ServerInfo struct {
	widget.BaseWidget
	container   *fyne.Container
	client      *redis.Client
	window      fyne.Window
	dbSelector  *widget.Select
	onDBChanged func(db int)

	// Info labels
	versionLabel    *widget.Label
	modeLabel       *widget.Label
	osLabel         *widget.Label
	uptimeLabel     *widget.Label
	clientsLabel    *widget.Label
	memoryLabel     *widget.Label
	memoryPeakLabel *widget.Label
	totalKeysLabel  *widget.Label
	expiredLabel    *widget.Label
	hitsLabel       *widget.Label
	missesLabel     *widget.Label
	hitRateLabel    *widget.Label
	lastRefreshLabel *widget.Label
}

// NewServerInfo creates a new server info panel
func NewServerInfo(window fyne.Window) *ServerInfo {
	si := &ServerInfo{
		window: window,
	}
	si.ExtendBaseWidget(si)
	si.buildUI()
	return si
}

func (si *ServerInfo) buildUI() {
	// Database selector
	dbOptions := make([]string, 16)
	for i := 0; i < 16; i++ {
		dbOptions[i] = fmt.Sprintf("DB %d", i)
	}
	si.dbSelector = widget.NewSelect(dbOptions, func(s string) {
		var db int
		fmt.Sscanf(s, "DB %d", &db)
		if si.onDBChanged != nil {
			si.onDBChanged(db)
		}
	})
	si.dbSelector.SetSelectedIndex(0)

	// Info labels
	si.versionLabel = widget.NewLabel("-")
	si.modeLabel = widget.NewLabel("-")
	si.osLabel = widget.NewLabel("-")
	si.uptimeLabel = widget.NewLabel("-")
	si.clientsLabel = widget.NewLabel("-")
	si.memoryLabel = widget.NewLabel("-")
	si.memoryPeakLabel = widget.NewLabel("-")
	si.totalKeysLabel = widget.NewLabel("-")
	si.expiredLabel = widget.NewLabel("-")
	si.hitsLabel = widget.NewLabel("-")
	si.missesLabel = widget.NewLabel("-")
	si.hitRateLabel = widget.NewLabel("-")
	si.lastRefreshLabel = widget.NewLabelWithStyle("-", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		si.Refresh()
	})

	// Server section
	serverSection := container.NewVBox(
		widget.NewLabelWithStyle("Server", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			widget.NewLabel("Version:"), si.versionLabel,
			widget.NewLabel("Mode:"), si.modeLabel,
			widget.NewLabel("OS:"), si.osLabel,
			widget.NewLabel("Uptime:"), si.uptimeLabel,
		),
	)

	// Clients section
	clientsSection := container.NewVBox(
		widget.NewLabelWithStyle("Clients", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			widget.NewLabel("Connected:"), si.clientsLabel,
		),
	)

	// Memory section
	memorySection := container.NewVBox(
		widget.NewLabelWithStyle("Memory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			widget.NewLabel("Used:"), si.memoryLabel,
			widget.NewLabel("Peak:"), si.memoryPeakLabel,
		),
	)

	// Keyspace section
	keyspaceSection := container.NewVBox(
		widget.NewLabelWithStyle("Keyspace", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			widget.NewLabel("Total Keys:"), si.totalKeysLabel,
			widget.NewLabel("Expired:"), si.expiredLabel,
			widget.NewLabel("Hits:"), si.hitsLabel,
			widget.NewLabel("Misses:"), si.missesLabel,
			widget.NewLabel("Hit Rate:"), si.hitRateLabel,
		),
	)

	// Database section
	dbSection := container.NewVBox(
		widget.NewLabelWithStyle("Database", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		si.dbSelector,
	)

	header := container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Server Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(si.lastRefreshLabel, refreshBtn),
	)

	content := container.NewVBox(
		serverSection,
		widget.NewSeparator(),
		clientsSection,
		widget.NewSeparator(),
		memorySection,
		widget.NewSeparator(),
		keyspaceSection,
		widget.NewSeparator(),
		dbSection,
	)

	scroll := container.NewVScroll(content)

	si.container = container.NewBorder(header, nil, nil, nil, scroll)
}

// CreateRenderer implements fyne.Widget
func (si *ServerInfo) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(si.container)
}

// SetClient sets the Redis client
func (si *ServerInfo) SetClient(client *redis.Client) {
	si.client = client
	if client != nil {
		// Update database selector with actual count from server
		dbCount := client.GetDatabaseCount()
		dbOptions := make([]string, dbCount)
		for i := 0; i < dbCount; i++ {
			dbOptions[i] = fmt.Sprintf("DB %d", i)
		}
		si.dbSelector.Options = dbOptions
		si.dbSelector.Refresh()
	}
}

// SetOnDBChanged sets the callback for database change
func (si *ServerInfo) SetOnDBChanged(f func(db int)) {
	si.onDBChanged = f
}

// Refresh updates the server info display
func (si *ServerInfo) Refresh() {
	if si.client == nil {
		si.clearInfo()
		return
	}

	info, err := si.client.GetServerInfo()
	if err != nil {
		si.clearInfo()
		si.lastRefreshLabel.SetText("Error: " + err.Error())
		return
	}

	si.versionLabel.SetText(info.Version)
	si.modeLabel.SetText(info.Mode)
	si.osLabel.SetText(info.OS)
	si.uptimeLabel.SetText(si.formatUptime(info.Uptime))
	si.clientsLabel.SetText(fmt.Sprintf("%d", info.ConnectedClients))
	si.memoryLabel.SetText(info.UsedMemoryHuman)
	si.memoryPeakLabel.SetText(si.formatBytes(info.UsedMemoryPeak))
	si.totalKeysLabel.SetText(fmt.Sprintf("%d", info.TotalKeys))
	si.expiredLabel.SetText(fmt.Sprintf("%d", info.ExpiredKeys))
	si.hitsLabel.SetText(fmt.Sprintf("%d", info.KeyspaceHits))
	si.missesLabel.SetText(fmt.Sprintf("%d", info.KeyspaceMisses))

	// Calculate hit rate
	totalOps := info.KeyspaceHits + info.KeyspaceMisses
	if totalOps > 0 {
		hitRate := float64(info.KeyspaceHits) / float64(totalOps) * 100
		si.hitRateLabel.SetText(fmt.Sprintf("%.2f%%", hitRate))
	} else {
		si.hitRateLabel.SetText("N/A")
	}

	// Update refresh timestamp
	si.lastRefreshLabel.SetText("Updated: " + time.Now().Format("15:04:05"))
}

func (si *ServerInfo) clearInfo() {
	si.versionLabel.SetText("-")
	si.modeLabel.SetText("-")
	si.osLabel.SetText("-")
	si.uptimeLabel.SetText("-")
	si.clientsLabel.SetText("-")
	si.memoryLabel.SetText("-")
	si.memoryPeakLabel.SetText("-")
	si.totalKeysLabel.SetText("-")
	si.expiredLabel.SetText("-")
	si.hitsLabel.SetText("-")
	si.missesLabel.SetText("-")
	si.hitRateLabel.SetText("-")
	si.lastRefreshLabel.SetText("-")
}

func (si *ServerInfo) formatUptime(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func (si *ServerInfo) formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// Clear clears the server info
func (si *ServerInfo) Clear() {
	si.clearInfo()
	si.dbSelector.SetSelectedIndex(0)
}
