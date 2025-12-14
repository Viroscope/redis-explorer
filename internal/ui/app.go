package ui

import (
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"redis-explorer/internal/config"
	"redis-explorer/internal/models"
	"redis-explorer/internal/redis"
)

// App represents the main application
type App struct {
	fyneApp       fyne.App
	window        fyne.Window
	sidebar       *Sidebar
	keyBrowser    *KeyBrowser
	editor        *ValueEditor
	serverInfo    *ServerInfo
	client        *redis.Client
	connected     bool
	currentDB     int
	appIcon       fyne.Resource
	refreshTicker *time.Ticker
	stopRefresh   chan struct{}
}

// NewApp creates a new application instance
func NewApp() *App {
	return &App{}
}

// Run starts the application
func (a *App) Run() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Create Fyne app
	a.fyneApp = app.NewWithID("com.redis-explorer")
	a.fyneApp.Settings().SetTheme(GetTheme(cfg.Theme))

	// Load app icon
	a.loadIcon()

	// Create main window
	a.window = a.fyneApp.NewWindow(AppName)
	a.window.Resize(fyne.NewSize(cfg.WindowWidth, cfg.WindowHeight))
	if a.appIcon != nil {
		a.window.SetIcon(a.appIcon)
	}

	// Create UI components
	a.createUI()

	// Set up window close handler
	a.window.SetOnClosed(func() {
		if a.connected {
			a.disconnect()
		}
		size := a.window.Canvas().Size()
		config.SetWindowSize(size.Width, size.Height)
	})

	// Show and run
	a.window.ShowAndRun()
}

func (a *App) createUI() {
	// Create components
	a.sidebar = NewSidebar(a.window)
	a.keyBrowser = NewKeyBrowser(a.window)
	a.editor = NewValueEditor(a.window)
	a.serverInfo = NewServerInfo(a.window)

	// Set up callbacks
	a.sidebar.SetOnConnect(func(conn models.ServerConnection) {
		a.connect(conn)
	})

	a.sidebar.SetOnDisconnect(func() {
		a.disconnect()
	})

	a.keyBrowser.SetOnKeySelected(func(key models.RedisKey) {
		a.editor.LoadKey(key)
	})

	a.keyBrowser.SetOnKeyDeleted(func(key string) {
		a.editor.Clear()
	})

	a.editor.SetOnKeyUpdated(func() {
		a.keyBrowser.LoadKeys()
	})

	a.serverInfo.SetOnDBChanged(func(db int) {
		a.selectDatabase(db)
	})

	// Create menu
	menu := a.createMenu()
	a.window.SetMainMenu(menu)

	// Create tabs for right panel
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Editor", theme.DocumentCreateIcon(), a.editor),
		container.NewTabItemWithIcon("Server Info", theme.InfoIcon(), a.serverInfo),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Main content: keys browser | editor/info tabs
	mainSplit := container.NewHSplit(a.keyBrowser, tabs)
	mainSplit.SetOffset(0.35)

	// Full layout: sidebar | main content
	fullSplit := container.NewHSplit(a.sidebar, mainSplit)
	fullSplit.SetOffset(0.18)

	a.window.SetContent(fullSplit)
}

func (a *App) createMenu() *fyne.MainMenu {
	// File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Settings", func() {
			ShowSettingsDialog(a.window, func() {
				// Restart auto-refresh with new settings
				if a.connected {
					a.stopAutoRefresh()
					a.startAutoRefresh()
				}
			})
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			a.fyneApp.Quit()
		}),
	)

	// View menu
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Theme", func() {
			cfg := config.Get()
			ShowThemeDialog(a.window, cfg.Theme, func(theme models.ThemeName) {
				config.SetTheme(theme)
				a.fyneApp.Settings().SetTheme(GetTheme(theme))
			})
		}),
		fyne.NewMenuItem("Refresh Keys", func() {
			if a.connected {
				a.keyBrowser.LoadKeys()
			}
		}),
	)

	// Connection menu
	connMenu := fyne.NewMenu("Connection",
		fyne.NewMenuItem("New Connection", func() {
			ShowConnectionDialog(a.window, nil, func(conn models.ServerConnection) {
				config.AddConnection(conn)
				a.sidebar.RefreshConnections()
			})
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Disconnect", func() {
			a.disconnect()
		}),
	)

	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			ShowAboutDialog(a.window, a.appIcon)
		}),
	)

	return fyne.NewMainMenu(fileMenu, viewMenu, connMenu, helpMenu)
}

func (a *App) connect(conn models.ServerConnection) {
	// Disconnect existing connection
	if a.connected {
		a.disconnect()
	}

	// Create new client
	a.client = redis.New(&conn)
	err := a.client.Connect()
	if err != nil {
		ShowErrorDialog(a.window, "Connection Error", err)
		return
	}

	a.connected = true
	a.currentDB = conn.Database

	// Update UI
	a.sidebar.SetConnected(true, conn.Name)
	a.keyBrowser.SetClient(a.client)
	a.editor.SetClient(a.client)
	a.serverInfo.SetClient(a.client)

	// Load data
	a.keyBrowser.LoadKeys()
	a.serverInfo.Refresh()

	// Start auto-refresh if configured
	a.startAutoRefresh()

	// Save last connection
	config.SetLastConnection(conn.ID)
}

func (a *App) disconnect() {
	if !a.connected {
		return
	}

	// Stop auto-refresh
	a.stopAutoRefresh()

	if a.client != nil {
		a.client.Disconnect()
		a.client = nil
	}

	a.connected = false

	// Clear UI
	a.sidebar.SetConnected(false, "")
	a.keyBrowser.SetClient(nil)
	a.keyBrowser.Clear()
	a.editor.SetClient(nil)
	a.editor.Clear()
	a.serverInfo.SetClient(nil)
	a.serverInfo.Clear()
}

func (a *App) selectDatabase(db int) {
	if !a.connected || a.client == nil {
		return
	}

	err := a.client.SelectDatabase(db)
	if err != nil {
		ShowErrorDialog(a.window, "Error", err)
		return
	}

	a.currentDB = db
	a.keyBrowser.LoadKeys()
	a.editor.Clear()
}

func (a *App) loadIcon() {
	// Try to load icon from various locations
	locations := []string{
		"icon.png",
		filepath.Join(".", "icon.png"),
	}

	// Get executable path and check there too
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		locations = append(locations, filepath.Join(execDir, "icon.png"))
	}

	for _, path := range locations {
		if data, err := os.ReadFile(path); err == nil {
			a.appIcon = fyne.NewStaticResource("icon.png", data)
			return
		}
	}
}

// startAutoRefresh starts the auto-refresh ticker if configured
func (a *App) startAutoRefresh() {
	cfg := config.Get()
	if cfg.AutoRefreshSecs <= 0 {
		return
	}

	a.stopRefresh = make(chan struct{})
	a.refreshTicker = time.NewTicker(time.Duration(cfg.AutoRefreshSecs) * time.Second)

	go func() {
		for {
			select {
			case <-a.refreshTicker.C:
				if a.connected {
					// Update UI on main thread (silent to avoid loading bar)
					fyne.Do(func() {
						a.keyBrowser.LoadKeysSilent()
						a.serverInfo.Refresh()
					})
				}
			case <-a.stopRefresh:
				return
			}
		}
	}()
}

// stopAutoRefresh stops the auto-refresh ticker
func (a *App) stopAutoRefresh() {
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
		a.refreshTicker = nil
	}
	if a.stopRefresh != nil {
		close(a.stopRefresh)
		a.stopRefresh = nil
	}
}
