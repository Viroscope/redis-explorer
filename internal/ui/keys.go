package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"redis-explorer/internal/models"
	"redis-explorer/internal/redis"
)

// TreeNode represents a node in the key tree
type TreeNode struct {
	ID       string
	Name     string
	FullKey  string
	IsKey    bool
	KeyType  string
	Children map[string]*TreeNode
}

// KeyBrowser represents the key browser panel
type KeyBrowser struct {
	widget.BaseWidget
	container     *fyne.Container
	contentArea   *fyne.Container
	keyList       *widget.List
	keyTree       *widget.Tree
	keys          []models.RedisKey
	filteredKeys  []models.RedisKey
	searchEntry   *widget.Entry
	typeFilter    *widget.Select
	countLabel    *widget.Label
	scopeLabel    *widget.Label
	clearScopeBtn *widget.Button
	setScopeBtn   *widget.Button
	client        *redis.Client
	onKeySelected func(key models.RedisKey)
	onKeyDeleted  func(key string)
	window        fyne.Window
	selectedIndex int
	selectedKey   string
	treeView      bool
	viewToggle    *widget.Button
	treeRoot      *TreeNode
	treeNodes     map[string]*TreeNode
	delimiter     string
	currentScope  string
	debounceTimer *time.Timer
	loadingBar    *widget.ProgressBarInfinite
	isLoading     bool
}

// NewKeyBrowser creates a new key browser panel
func NewKeyBrowser(window fyne.Window) *KeyBrowser {
	kb := &KeyBrowser{
		window:        window,
		selectedIndex: -1,
		treeView:      false,
		delimiter:     ":",
		treeNodes:     make(map[string]*TreeNode),
		currentScope:  "",
	}
	kb.ExtendBaseWidget(kb)
	kb.buildUI()
	return kb
}

func (kb *KeyBrowser) buildUI() {
	// Count label (must be created first as filterKeys uses it)
	kb.countLabel = widget.NewLabel("0 keys")

	// Scope indicator and controls
	kb.scopeLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	kb.clearScopeBtn = widget.NewButtonWithIcon("Clear", theme.CancelIcon(), func() {
		kb.clearScope()
	})
	kb.clearScopeBtn.Importance = widget.LowImportance
	kb.clearScopeBtn.Hide()

	kb.setScopeBtn = widget.NewButtonWithIcon("Scope", theme.SearchIcon(), func() {
		kb.setScopeFromSelection()
	})
	kb.setScopeBtn.Importance = widget.LowImportance

	// Search entry with debouncing
	kb.searchEntry = widget.NewEntry()
	kb.searchEntry.SetPlaceHolder("Search keys...")
	kb.searchEntry.OnChanged = func(s string) {
		// Debounce search to avoid excessive filtering on each keystroke
		if kb.debounceTimer != nil {
			kb.debounceTimer.Stop()
		}
		kb.debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
			// Update UI on main thread
			fyne.Do(func() {
				kb.filterKeys()
			})
		})
	}

	// Type filter
	kb.typeFilter = widget.NewSelect([]string{"All Types", "string", "list", "set", "hash", "zset", "stream"}, func(s string) {
		kb.filterKeys()
	})
	kb.typeFilter.SetSelected("All Types")

	// Build list view
	kb.keyList = kb.buildListView()

	// Build tree view
	kb.keyTree = kb.buildTreeView()

	// Content area that holds either list or tree
	kb.contentArea = container.NewStack(kb.keyList)

	// Loading indicator
	kb.loadingBar = widget.NewProgressBarInfinite()
	kb.loadingBar.Hide()

	// View toggle button
	kb.viewToggle = widget.NewButtonWithIcon("View", theme.ListIcon(), func() {
		kb.toggleView()
	})
	kb.viewToggle.Importance = widget.LowImportance

	// Buttons
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		kb.LoadKeys()
	})
	refreshBtn.Importance = widget.LowImportance

	newKeyBtn := widget.NewButtonWithIcon("New", theme.ContentAddIcon(), func() {
		if kb.client == nil {
			return
		}
		ShowNewKeyDialog(kb.window, func(key string, keyType string) {
			kb.createKey(key, keyType)
		})
	})
	newKeyBtn.Importance = widget.LowImportance

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		kb.deleteSelectedKey()
	})
	deleteBtn.Importance = widget.LowImportance

	// Search bar with filter
	searchBar := container.NewBorder(nil, nil, nil,
		kb.typeFilter,
		kb.searchEntry,
	)

	// Scope bar
	scopeBar := container.NewHBox(kb.scopeLabel, kb.clearScopeBtn)

	// Button bar with view toggle
	buttonBar := container.NewHBox(
		kb.viewToggle,
		widget.NewSeparator(),
		refreshBtn,
		newKeyBtn,
		deleteBtn,
		widget.NewSeparator(),
		kb.setScopeBtn,
	)

	// Header
	header := container.NewVBox(
		container.NewBorder(nil, nil,
			widget.NewLabelWithStyle("Keys", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			kb.countLabel,
			nil,
		),
		scopeBar,
		searchBar,
		buttonBar,
		kb.loadingBar,
	)

	kb.container = container.NewBorder(header, nil, nil, nil, kb.contentArea)
}

func (kb *KeyBrowser) buildListView() *widget.List {
	list := widget.NewList(
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
			icon.SetResource(kb.getKeyIcon(key.Type))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		kb.selectedIndex = id
		if kb.onKeySelected != nil && id >= 0 && id < len(kb.filteredKeys) {
			kb.selectedKey = kb.filteredKeys[id].Key
			kb.onKeySelected(kb.filteredKeys[id])
		}
	}

	return list
}

func (kb *KeyBrowser) buildTreeView() *widget.Tree {
	tree := widget.NewTree(
		// ChildUIDs - returns child IDs for a node
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if uid == "" {
				// Root level
				if kb.treeRoot == nil {
					return []widget.TreeNodeID{}
				}
				return kb.getChildIDs(kb.treeRoot)
			}
			node, ok := kb.treeNodes[uid]
			if !ok || node == nil {
				return []widget.TreeNodeID{}
			}
			return kb.getChildIDs(node)
		},
		// IsBranch - returns true if the node has children
		func(uid widget.TreeNodeID) bool {
			if uid == "" {
				return true
			}
			node, ok := kb.treeNodes[uid]
			if !ok || node == nil {
				return false
			}
			return len(node.Children) > 0
		},
		// CreateNode - creates a new node widget
		func(branch bool) fyne.CanvasObject {
			label := widget.NewLabel("Node")
			icon := widget.NewIcon(theme.FolderIcon())
			typeLabel := widget.NewLabel("")
			row := container.NewHBox(icon, label, typeLabel)
			return row
		},
		// UpdateNode - updates the node widget
		func(uid widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			node, ok := kb.treeNodes[uid]
			if !ok || node == nil {
				return
			}

			box := o.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			nameLabel := box.Objects[1].(*widget.Label)
			typeLabel := box.Objects[2].(*widget.Label)

			nameLabel.SetText(node.Name)

			if node.IsKey {
				icon.SetResource(kb.getKeyIcon(node.KeyType))
				typeLabel.SetText(fmt.Sprintf("[%s]", node.KeyType))
			} else {
				icon.SetResource(theme.FolderIcon())
				// Count keys in this folder
				count := kb.countKeysInNode(node)
				typeLabel.SetText(fmt.Sprintf("(%d)", count))
			}
		},
	)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		node, ok := kb.treeNodes[uid]
		if !ok || node == nil {
			return
		}

		kb.selectedKey = uid

		if node.IsKey {
			// Find the key in filteredKeys
			for _, key := range kb.filteredKeys {
				if key.Key == node.FullKey {
					if kb.onKeySelected != nil {
						kb.onKeySelected(key)
					}
					break
				}
			}
		}
	}

	return tree
}

func (kb *KeyBrowser) setScopeFromSelection() {
	var scopePath string

	if kb.treeView {
		// In tree view, use the selected node
		if kb.selectedKey == "" {
			return
		}
		if node, ok := kb.treeNodes[kb.selectedKey]; ok {
			if node.IsKey {
				// It's a key - get the parent path
				lastDelim := strings.LastIndex(kb.selectedKey, kb.delimiter)
				if lastDelim > 0 {
					scopePath = kb.selectedKey[:lastDelim]
				}
			} else {
				// It's a folder - use it directly
				scopePath = kb.selectedKey
			}
		}
	} else {
		// In list view, extract prefix from selected key
		if kb.selectedIndex >= 0 && kb.selectedIndex < len(kb.filteredKeys) {
			key := kb.filteredKeys[kb.selectedIndex].Key
			lastDelim := strings.LastIndex(key, kb.delimiter)
			if lastDelim > 0 {
				scopePath = key[:lastDelim]
			}
		}
	}

	if scopePath != "" {
		kb.setScope(scopePath)
	}
}

func (kb *KeyBrowser) setScope(scope string) {
	kb.currentScope = scope
	kb.scopeLabel.SetText("Scope: " + scope)
	kb.clearScopeBtn.Show()
	kb.filterKeys()
}

func (kb *KeyBrowser) clearScope() {
	kb.currentScope = ""
	kb.scopeLabel.SetText("")
	kb.clearScopeBtn.Hide()
	kb.filterKeys()
}

func (kb *KeyBrowser) getChildIDs(node *TreeNode) []widget.TreeNodeID {
	var ids []widget.TreeNodeID
	for _, child := range node.Children {
		ids = append(ids, child.ID)
	}
	// Sort for consistent order
	sort.Strings(ids)
	return ids
}

func (kb *KeyBrowser) countKeysInNode(node *TreeNode) int {
	count := 0
	if node.IsKey {
		count = 1
	}
	for _, child := range node.Children {
		count += kb.countKeysInNode(child)
	}
	return count
}

func (kb *KeyBrowser) getKeyIcon(keyType string) fyne.Resource {
	switch keyType {
	case "string":
		return theme.DocumentIcon()
	case "list":
		return theme.ListIcon()
	case "set":
		return theme.GridIcon()
	case "hash":
		return theme.StorageIcon()
	case "zset":
		return theme.MenuIcon()
	default:
		return theme.FileIcon()
	}
}

func (kb *KeyBrowser) toggleView() {
	kb.treeView = !kb.treeView
	kb.contentArea.RemoveAll()

	if kb.treeView {
		kb.viewToggle.SetIcon(theme.FolderIcon())
		kb.buildKeyTree()
		kb.contentArea.Add(kb.keyTree)
		kb.keyTree.Refresh()
	} else {
		kb.viewToggle.SetIcon(theme.ListIcon())
		kb.contentArea.Add(kb.keyList)
		kb.keyList.Refresh()
	}
	kb.contentArea.Refresh()
}

func (kb *KeyBrowser) buildKeyTree() {
	kb.treeNodes = make(map[string]*TreeNode)
	kb.treeRoot = &TreeNode{
		ID:       "",
		Name:     "root",
		Children: make(map[string]*TreeNode),
	}

	for _, key := range kb.filteredKeys {
		kb.addKeyToTree(key)
	}
}

func (kb *KeyBrowser) addKeyToTree(key models.RedisKey) {
	parts := strings.Split(key.Key, kb.delimiter)
	currentNode := kb.treeRoot
	currentPath := ""

	for i, part := range parts {
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + kb.delimiter + part
		}

		isLastPart := i == len(parts)-1

		child, exists := currentNode.Children[part]
		if !exists {
			child = &TreeNode{
				ID:       currentPath,
				Name:     part,
				FullKey:  key.Key,
				IsKey:    isLastPart,
				KeyType:  key.Type,
				Children: make(map[string]*TreeNode),
			}
			currentNode.Children[part] = child
			kb.treeNodes[currentPath] = child
		}

		if isLastPart {
			// Mark as key if this is the final part
			child.IsKey = true
			child.FullKey = key.Key
			child.KeyType = key.Type
		}

		currentNode = child
	}
}

func (kb *KeyBrowser) deleteSelectedKey() {
	var keyToDelete string

	if kb.treeView {
		keyToDelete = kb.selectedKey
		// Check if it's actually a key (not a folder)
		if node, ok := kb.treeNodes[keyToDelete]; ok && !node.IsKey {
			return // Can't delete a folder
		}
	} else {
		if kb.selectedIndex < 0 || kb.selectedIndex >= len(kb.filteredKeys) {
			return
		}
		keyToDelete = kb.filteredKeys[kb.selectedIndex].Key
	}

	if keyToDelete == "" {
		return
	}

	ShowConfirmDialog(kb.window, "Delete Key",
		fmt.Sprintf("Are you sure you want to delete '%s'?", keyToDelete),
		func() {
			if kb.client != nil {
				err := kb.client.DeleteKey(keyToDelete)
				if err != nil {
					ShowErrorDialog(kb.window, "Error", err)
					return
				}
				if kb.onKeyDeleted != nil {
					kb.onKeyDeleted(keyToDelete)
				}
				kb.LoadKeys()
			}
		})
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
		// Scope filter - key must start with scope prefix
		if kb.currentScope != "" {
			if !strings.HasPrefix(key.Key, kb.currentScope+kb.delimiter) && key.Key != kb.currentScope {
				continue
			}
		}

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

	if kb.treeView {
		kb.buildKeyTree()
		if kb.keyTree != nil {
			kb.keyTree.Refresh()
		}
	} else {
		if kb.keyList != nil {
			kb.keyList.Refresh()
		}
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

// LoadKeys loads keys from the connected Redis server asynchronously
func (kb *KeyBrowser) LoadKeys() {
	kb.loadKeysInternal(false)
}

// LoadKeysSilent loads keys without showing the loading bar (for auto-refresh)
func (kb *KeyBrowser) LoadKeysSilent() {
	kb.loadKeysInternal(true)
}

func (kb *KeyBrowser) loadKeysInternal(silent bool) {
	if kb.client == nil {
		kb.keys = nil
		kb.filteredKeys = nil
		if kb.countLabel != nil {
			kb.countLabel.SetText("0 keys")
		}
		if kb.keyList != nil {
			kb.keyList.Refresh()
		}
		if kb.keyTree != nil {
			kb.buildKeyTree()
			kb.keyTree.Refresh()
		}
		return
	}

	// Prevent multiple concurrent loads
	if kb.isLoading {
		return
	}

	kb.isLoading = true
	if !silent {
		kb.loadingBar.Show()
		kb.loadingBar.Start()
		if kb.countLabel != nil {
			kb.countLabel.SetText("Loading...")
		}
	}

	// Load keys in background goroutine
	go func() {
		keys, err := kb.client.GetAllKeys("*", 10000)

		// Update UI on main thread using fyne.Do
		fyne.Do(func() {
			kb.isLoading = false
			if !silent {
				kb.loadingBar.Stop()
				kb.loadingBar.Hide()
			}

			if err != nil {
				if kb.countLabel != nil {
					kb.countLabel.SetText("Error")
				}
				if !silent {
					ShowErrorDialog(kb.window, "Error loading keys", err)
				}
				return
			}

			kb.keys = keys
			kb.filterKeys()
		})
	}()
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
	kb.selectedKey = ""
	kb.clearScope()
	if kb.countLabel != nil {
		kb.countLabel.SetText("0 keys")
	}
	kb.selectedIndex = -1
	if kb.keyList != nil {
		kb.keyList.UnselectAll()
		kb.keyList.Refresh()
	}
	if kb.keyTree != nil {
		kb.buildKeyTree()
		kb.keyTree.Refresh()
	}
}

// GetSelectedKey returns the currently selected key
func (kb *KeyBrowser) GetSelectedKey() *models.RedisKey {
	if kb.treeView {
		for i := range kb.filteredKeys {
			if kb.filteredKeys[i].Key == kb.selectedKey {
				return &kb.filteredKeys[i]
			}
		}
		return nil
	}
	if kb.selectedIndex >= 0 && kb.selectedIndex < len(kb.filteredKeys) {
		return &kb.filteredKeys[kb.selectedIndex]
	}
	return nil
}
