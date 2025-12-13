package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"redis-explorer/internal/models"
	"redis-explorer/internal/redis"
)

// KeyBrowser represents the key browser panel
type KeyBrowser struct {
	widget.BaseWidget
	container     *fyne.Container
	keyList       *widget.List
	keys          []models.RedisKey
	filteredKeys  []models.RedisKey
	searchEntry   *widget.Entry
	typeFilter    *widget.Select
	countLabel    *widget.Label
	client        *redis.Client
	onKeySelected func(key models.RedisKey)
	onKeyDeleted  func(key string)
	window        fyne.Window
	selectedIndex int
}

// NewKeyBrowser creates a new key browser panel
func NewKeyBrowser(window fyne.Window) *KeyBrowser {
	kb := &KeyBrowser{
		window:        window,
		selectedIndex: -1,
	}
	kb.ExtendBaseWidget(kb)
	kb.buildUI()
	return kb
}

func (kb *KeyBrowser) buildUI() {
	// Count label (must be created first as filterKeys uses it)
	kb.countLabel = widget.NewLabel("0 keys")

	// Search entry
	kb.searchEntry = widget.NewEntry()
	kb.searchEntry.SetPlaceHolder("Search keys (pattern: *)...")
	kb.searchEntry.OnChanged = func(s string) {
		kb.filterKeys()
	}

	// Type filter
	kb.typeFilter = widget.NewSelect([]string{"All Types", "string", "list", "set", "hash", "zset", "stream"}, func(s string) {
		kb.filterKeys()
	})
	kb.typeFilter.SetSelected("All Types")

	// Key list
	kb.keyList = widget.NewList(
		func() int { return len(kb.filteredKeys) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Key Name"),
				widget.NewLabel("[type]"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			nameLabel := box.Objects[1].(*widget.Label)
			typeLabel := box.Objects[2].(*widget.Label)

			key := kb.filteredKeys[i]
			nameLabel.SetText(key.Key)
			typeLabel.SetText(fmt.Sprintf("[%s]", key.Type))

			// Set icon based on type
			switch key.Type {
			case "string":
				icon.SetResource(theme.DocumentIcon())
			case "list":
				icon.SetResource(theme.ListIcon())
			case "set":
				icon.SetResource(theme.GridIcon())
			case "hash":
				icon.SetResource(theme.StorageIcon())
			case "zset":
				icon.SetResource(theme.MenuIcon())
			default:
				icon.SetResource(theme.FileIcon())
			}
		},
	)

	kb.keyList.OnSelected = func(id widget.ListItemID) {
		kb.selectedIndex = id
		if kb.onKeySelected != nil && id >= 0 && id < len(kb.filteredKeys) {
			kb.onKeySelected(kb.filteredKeys[id])
		}
	}

	// Buttons
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		kb.LoadKeys()
	})

	newKeyBtn := widget.NewButtonWithIcon("New Key", theme.ContentAddIcon(), func() {
		if kb.client == nil {
			return
		}
		ShowNewKeyDialog(kb.window, func(key string, keyType string) {
			kb.createKey(key, keyType)
		})
	})

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if kb.selectedIndex < 0 || kb.selectedIndex >= len(kb.filteredKeys) {
			return
		}
		key := kb.filteredKeys[kb.selectedIndex]
		ShowConfirmDialog(kb.window, "Delete Key",
			fmt.Sprintf("Are you sure you want to delete '%s'?", key.Key),
			func() {
				if kb.client != nil {
					err := kb.client.DeleteKey(key.Key)
					if err != nil {
						ShowErrorDialog(kb.window, "Error", err)
						return
					}
					if kb.onKeyDeleted != nil {
						kb.onKeyDeleted(key.Key)
					}
					kb.LoadKeys()
				}
			})
	})

	// Search bar with filter
	searchBar := container.NewBorder(nil, nil, nil,
		container.NewHBox(kb.typeFilter),
		kb.searchEntry,
	)

	// Button bar
	buttonBar := container.NewHBox(refreshBtn, newKeyBtn, deleteBtn)

	// Header
	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Keys", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			kb.countLabel,
		),
		searchBar,
		buttonBar,
	)

	kb.container = container.NewBorder(header, nil, nil, nil, kb.keyList)
}

func (kb *KeyBrowser) createKey(key string, keyType string) {
	if kb.client == nil {
		return
	}

	var err error
	switch keyType {
	case "string":
		err = kb.client.SetString(key, "")
	case "list":
		err = kb.client.ListPush(key, "", false)
	case "set":
		err = kb.client.SetAdd(key, "")
	case "hash":
		err = kb.client.HashSet(key, "field", "")
	case "zset":
		err = kb.client.SortedSetAdd(key, 0, "")
	}

	if err != nil {
		ShowErrorDialog(kb.window, "Error", err)
		return
	}

	kb.LoadKeys()
}

func (kb *KeyBrowser) filterKeys() {
	var pattern string
	var typeFilter string

	if kb.searchEntry != nil {
		pattern = strings.ToLower(kb.searchEntry.Text)
	}
	if kb.typeFilter != nil {
		typeFilter = kb.typeFilter.Selected
	}

	kb.filteredKeys = nil
	for _, key := range kb.keys {
		// Type filter
		if typeFilter != "" && typeFilter != "All Types" && key.Type != typeFilter {
			continue
		}

		// Search filter
		if pattern != "" && !strings.Contains(strings.ToLower(key.Key), pattern) {
			continue
		}

		kb.filteredKeys = append(kb.filteredKeys, key)
	}

	if kb.countLabel != nil {
		kb.countLabel.SetText(fmt.Sprintf("%d keys", len(kb.filteredKeys)))
	}
	if kb.keyList != nil {
		kb.keyList.Refresh()
	}
}

// CreateRenderer implements fyne.Widget
func (kb *KeyBrowser) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(kb.container)
}

// SetClient sets the Redis client
func (kb *KeyBrowser) SetClient(client *redis.Client) {
	kb.client = client
}

// LoadKeys loads keys from the connected Redis server
func (kb *KeyBrowser) LoadKeys() {
	if kb.client == nil {
		kb.keys = nil
		kb.filteredKeys = nil
		kb.countLabel.SetText("0 keys")
		kb.keyList.Refresh()
		return
	}

	keys, err := kb.client.GetAllKeys("*", 10000)
	if err != nil {
		ShowErrorDialog(kb.window, "Error loading keys", err)
		return
	}

	kb.keys = keys
	kb.filterKeys()
}

// SetOnKeySelected sets the callback for key selection
func (kb *KeyBrowser) SetOnKeySelected(f func(key models.RedisKey)) {
	kb.onKeySelected = f
}

// SetOnKeyDeleted sets the callback for key deletion
func (kb *KeyBrowser) SetOnKeyDeleted(f func(key string)) {
	kb.onKeyDeleted = f
}

// Clear clears the key list
func (kb *KeyBrowser) Clear() {
	kb.keys = nil
	kb.filteredKeys = nil
	kb.countLabel.SetText("0 keys")
	kb.selectedIndex = -1
	kb.keyList.UnselectAll()
	kb.keyList.Refresh()
}

// GetSelectedKey returns the currently selected key
func (kb *KeyBrowser) GetSelectedKey() *models.RedisKey {
	if kb.selectedIndex >= 0 && kb.selectedIndex < len(kb.filteredKeys) {
		return &kb.filteredKeys[kb.selectedIndex]
	}
	return nil
}
