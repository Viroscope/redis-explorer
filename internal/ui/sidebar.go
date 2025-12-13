package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"redis-explorer/internal/config"
	"redis-explorer/internal/models"
)

// Sidebar represents the connection sidebar panel
type Sidebar struct {
	widget.BaseWidget
	container    *fyne.Container
	connList     *widget.List
	connections  []models.ServerConnection
	selectedID   string
	onConnect    func(conn models.ServerConnection)
	onDisconnect func()
	onEdit       func(conn models.ServerConnection)
	onDelete     func(conn models.ServerConnection)
	window       fyne.Window
	isConnected  bool
	statusLabel  *widget.Label
}

// NewSidebar creates a new sidebar
func NewSidebar(window fyne.Window) *Sidebar {
	s := &Sidebar{
		window:      window,
		statusLabel: widget.NewLabel("Disconnected"),
	}
	s.ExtendBaseWidget(s)
	s.loadConnections()
	s.buildUI()
	return s
}

func (s *Sidebar) loadConnections() {
	cfg := config.Get()
	s.connections = cfg.Connections
	if cfg.LastConnectionID != "" {
		s.selectedID = cfg.LastConnectionID
	} else if len(s.connections) > 0 {
		s.selectedID = s.connections[0].ID
	}
}

func (s *Sidebar) buildUI() {
	// Connection list
	s.connList = widget.NewList(
		func() int { return len(s.connections) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.ComputerIcon()),
				widget.NewLabel("Connection Name"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			label.SetText(s.connections[i].Name)
		},
	)

	s.connList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(s.connections) {
			s.selectedID = s.connections[id].ID
		}
	}

	// Select the last used connection
	for i, conn := range s.connections {
		if conn.ID == s.selectedID {
			s.connList.Select(i)
			break
		}
	}

	// Buttons
	connectBtn := widget.NewButtonWithIcon("Connect", theme.LoginIcon(), func() {
		if s.onConnect != nil && s.selectedID != "" {
			for _, conn := range s.connections {
				if conn.ID == s.selectedID {
					s.onConnect(conn)
					break
				}
			}
		}
	})

	disconnectBtn := widget.NewButtonWithIcon("Disconnect", theme.LogoutIcon(), func() {
		if s.onDisconnect != nil {
			s.onDisconnect()
		}
	})

	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		ShowConnectionDialog(s.window, nil, func(conn models.ServerConnection) {
			config.AddConnection(conn)
			s.loadConnections()
			s.connList.Refresh()
		})
	})

	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		if s.selectedID == "" {
			return
		}
		for _, conn := range s.connections {
			if conn.ID == s.selectedID {
				ShowConnectionDialog(s.window, &conn, func(updated models.ServerConnection) {
					config.UpdateConnection(updated)
					s.loadConnections()
					s.connList.Refresh()
					if s.onEdit != nil {
						s.onEdit(updated)
					}
				})
				break
			}
		}
	})

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if s.selectedID == "" || s.selectedID == "default" {
			return
		}
		for _, conn := range s.connections {
			if conn.ID == s.selectedID {
				ShowConfirmDialog(s.window, "Delete Connection",
					fmt.Sprintf("Are you sure you want to delete '%s'?", conn.Name),
					func() {
						config.RemoveConnection(conn.ID)
						s.loadConnections()
						s.connList.Refresh()
						if s.onDelete != nil {
							s.onDelete(conn)
						}
					})
				break
			}
		}
	})

	buttonBar := container.NewHBox(addBtn, editBtn, deleteBtn)

	// Status
	statusContainer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			s.statusLabel,
		),
	)

	// Build the layout
	s.container = container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Connections", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			buttonBar,
		),
		container.NewVBox(
			statusContainer,
			widget.NewSeparator(),
			connectBtn,
			disconnectBtn,
		),
		nil, nil,
		s.connList,
	)
}

// CreateRenderer implements fyne.Widget
func (s *Sidebar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}

// SetOnConnect sets the connection callback
func (s *Sidebar) SetOnConnect(f func(conn models.ServerConnection)) {
	s.onConnect = f
}

// SetOnDisconnect sets the disconnection callback
func (s *Sidebar) SetOnDisconnect(f func()) {
	s.onDisconnect = f
}

// SetOnEdit sets the edit callback
func (s *Sidebar) SetOnEdit(f func(conn models.ServerConnection)) {
	s.onEdit = f
}

// SetOnDelete sets the delete callback
func (s *Sidebar) SetOnDelete(f func(conn models.ServerConnection)) {
	s.onDelete = f
}

// SetConnected updates the connection status display
func (s *Sidebar) SetConnected(connected bool, connName string) {
	s.isConnected = connected
	if connected {
		s.statusLabel.SetText(fmt.Sprintf("Connected: %s", connName))
	} else {
		s.statusLabel.SetText("Disconnected")
	}
}

// RefreshConnections reloads connections from config
func (s *Sidebar) RefreshConnections() {
	s.loadConnections()
	s.connList.Refresh()
}
