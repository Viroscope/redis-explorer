package ui

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"redis-explorer/internal/models"
	"redis-explorer/internal/redis"
)

// EditableLabel is a label that can be double-clicked to edit
type EditableLabel struct {
	widget.BaseWidget
	text      string
	label     *widget.Label
	onEdit    func(newValue string)
	lastClick time.Time
	window    fyne.Window
	fieldName string
}

// NewEditableLabel creates a new editable label
func NewEditableLabel(text string, fieldName string, window fyne.Window, onEdit func(newValue string)) *EditableLabel {
	el := &EditableLabel{
		text:      text,
		fieldName: fieldName,
		window:    window,
		onEdit:    onEdit,
	}
	el.ExtendBaseWidget(el)
	el.label = widget.NewLabel(text)
	return el
}

func (el *EditableLabel) Tapped(e *fyne.PointEvent) {
	now := time.Now()
	if now.Sub(el.lastClick) < 400*time.Millisecond {
		// Double click - show edit dialog
		el.showEditDialog()
	}
	el.lastClick = now
}

func (el *EditableLabel) showEditDialog() {
	entry := widget.NewEntry()
	entry.SetText(el.text)
	entry.MultiLine = false

	dialog.ShowForm(fmt.Sprintf("Edit %s", el.fieldName), "Save", "Cancel",
		[]*widget.FormItem{
			{Text: el.fieldName, Widget: entry},
		},
		func(save bool) {
			if save && el.onEdit != nil {
				el.onEdit(entry.Text)
			}
		}, el.window)
}

func (el *EditableLabel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(el.label)
}

func (el *EditableLabel) SetText(text string) {
	el.text = text
	el.label.SetText(text)
}

// ValueEditor represents the value editor panel
type ValueEditor struct {
	widget.BaseWidget
	container    *fyne.Container
	keyLabel     *widget.Label
	typeLabel    *widget.Label
	ttlLabel     *widget.Label
	contentArea  *fyne.Container
	client       *redis.Client
	currentKey   *models.RedisKey
	window       fyne.Window
	onKeyUpdated func()
}

// NewValueEditor creates a new value editor panel
func NewValueEditor(window fyne.Window) *ValueEditor {
	ve := &ValueEditor{
		window: window,
	}
	ve.ExtendBaseWidget(ve)
	ve.buildUI()
	return ve
}

func (ve *ValueEditor) buildUI() {
	ve.keyLabel = widget.NewLabelWithStyle("No key selected", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ve.typeLabel = widget.NewLabel("")
	ve.ttlLabel = widget.NewLabel("")

	ttlBtn := widget.NewButtonWithIcon("Set TTL", theme.HistoryIcon(), func() {
		if ve.currentKey == nil || ve.client == nil {
			return
		}
		ShowTTLDialog(ve.window, ve.currentKey.TTL, func(ttl int64) {
			err := ve.client.SetTTL(ve.currentKey.Key, ttl)
			if err != nil {
				ShowErrorDialog(ve.window, "Error", err)
				return
			}
			ve.refreshTTL()
		})
	})

	header := container.NewVBox(
		ve.keyLabel,
		container.NewHBox(ve.typeLabel, ve.ttlLabel, ttlBtn),
		widget.NewSeparator(),
	)

	ve.contentArea = container.NewMax(widget.NewLabel("Select a key to view its value"))

	ve.container = container.NewBorder(header, nil, nil, nil, ve.contentArea)
}

func (ve *ValueEditor) refreshTTL() {
	if ve.currentKey == nil || ve.client == nil {
		return
	}
	ttl, _ := ve.client.GetTTL(ve.currentKey.Key)
	ve.currentKey.TTL = ttl
	if ttl < 0 {
		ve.ttlLabel.SetText("TTL: No expiry")
	} else {
		ve.ttlLabel.SetText(fmt.Sprintf("TTL: %ds", ttl))
	}
}

// CreateRenderer implements fyne.Widget
func (ve *ValueEditor) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ve.container)
}

// SetClient sets the Redis client
func (ve *ValueEditor) SetClient(client *redis.Client) {
	ve.client = client
}

// SetOnKeyUpdated sets the callback for when a key is updated
func (ve *ValueEditor) SetOnKeyUpdated(f func()) {
	ve.onKeyUpdated = f
}

// LoadKey loads a key's value into the editor
func (ve *ValueEditor) LoadKey(key models.RedisKey) {
	ve.currentKey = &key
	ve.keyLabel.SetText(key.Key)
	ve.typeLabel.SetText(fmt.Sprintf("Type: %s", key.Type))

	if key.TTL < 0 {
		ve.ttlLabel.SetText("TTL: No expiry")
	} else {
		ve.ttlLabel.SetText(fmt.Sprintf("TTL: %ds", key.TTL))
	}

	ve.loadValueEditor(key)
}

func (ve *ValueEditor) loadValueEditor(key models.RedisKey) {
	if ve.client == nil {
		return
	}

	var content fyne.CanvasObject

	switch key.Type {
	case "string":
		content = ve.buildStringEditor(key)
	case "list":
		content = ve.buildListEditor(key)
	case "set":
		content = ve.buildSetEditor(key)
	case "hash":
		content = ve.buildHashEditor(key)
	case "zset":
		content = ve.buildZSetEditor(key)
	default:
		content = widget.NewLabel("Unsupported key type: " + key.Type)
	}

	ve.contentArea.RemoveAll()
	ve.contentArea.Add(content)
	ve.contentArea.Refresh()
}

func (ve *ValueEditor) buildStringEditor(key models.RedisKey) fyne.CanvasObject {
	value, err := ve.client.GetString(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	entry := widget.NewMultiLineEntry()
	entry.SetText(value)
	entry.Wrapping = fyne.TextWrapWord

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		err := ve.client.SetString(key.Key, entry.Text)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		ShowInfoDialog(ve.window, "Success", "Value saved")
		if ve.onKeyUpdated != nil {
			ve.onKeyUpdated()
		}
	})

	hint := widget.NewLabelWithStyle("Edit the value above and click Save", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	return container.NewBorder(nil, container.NewVBox(hint, saveBtn), nil, nil, entry)
}

func (ve *ValueEditor) buildListEditor(key models.RedisKey) fyne.CanvasObject {
	items, err := ve.client.GetList(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	// Build table-like grid with aligned columns
	table := widget.NewTable(
		func() (int, int) { return len(items), 2 },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{}),
			)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			if id.Col == 0 {
				label.SetText(fmt.Sprintf("[%d]", id.Row))
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.SetText(items[id.Row])
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 400)

	// Double-click to edit
	table.OnSelected = func(id widget.TableCellID) {
		if id.Col == 1 && id.Row < len(items) {
			ve.showEditValueDialog("Value", items[id.Row], func(newVal string) {
				err := ve.client.ListSet(key.Key, int64(id.Row), newVal)
				if err != nil {
					ShowErrorDialog(ve.window, "Error", err)
					return
				}
				ve.LoadKey(key)
			})
		}
		table.UnselectAll()
	}

	addEntry := widget.NewEntry()
	addEntry.SetPlaceHolder("New value")

	addLeftBtn := widget.NewButtonWithIcon("Add Left", theme.ContentAddIcon(), func() {
		if addEntry.Text == "" {
			return
		}
		err := ve.client.ListPush(key.Key, addEntry.Text, true)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		addEntry.SetText("")
		ve.LoadKey(key)
	})

	addRightBtn := widget.NewButtonWithIcon("Add Right", theme.ContentAddIcon(), func() {
		if addEntry.Text == "" {
			return
		}
		err := ve.client.ListPush(key.Key, addEntry.Text, false)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		addEntry.SetText("")
		ve.LoadKey(key)
	})

	hint := widget.NewLabelWithStyle("Click a value to edit", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	addBar := container.NewVBox(
		hint,
		container.NewBorder(nil, nil, nil,
			container.NewHBox(addLeftBtn, addRightBtn),
			addEntry,
		),
	)

	return container.NewBorder(nil, addBar, nil, nil, table)
}

func (ve *ValueEditor) buildSetEditor(key models.RedisKey) fyne.CanvasObject {
	members, err := ve.client.GetSet(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	sort.Strings(members)
	var selectedMember string
	var selectedRow int = -1

	table := widget.NewTable(
		func() (int, int) { return len(members), 1 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(members[id.Row])
		},
	)
	table.SetColumnWidth(0, 450)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(members) {
			selectedMember = members[id.Row]
			selectedRow = id.Row
		}
	}

	addEntry := widget.NewEntry()
	addEntry.SetPlaceHolder("New member")

	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		if addEntry.Text == "" {
			return
		}
		err := ve.client.SetAdd(key.Key, addEntry.Text)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		addEntry.SetText("")
		ve.LoadKey(key)
	})

	removeBtn := widget.NewButtonWithIcon("Remove Selected", theme.ContentRemoveIcon(), func() {
		if selectedMember == "" || selectedRow < 0 {
			return
		}
		err := ve.client.SetRemove(key.Key, selectedMember)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		selectedMember = ""
		selectedRow = -1
		ve.LoadKey(key)
	})

	addBar := container.NewVBox(
		container.NewBorder(nil, nil, nil, addBtn, addEntry),
		removeBtn,
	)

	return container.NewBorder(nil, addBar, nil, nil, table)
}

func (ve *ValueEditor) buildHashEditor(key models.RedisKey) fyne.CanvasObject {
	hash, err := ve.client.GetHash(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	// Convert map to sorted slice
	type fieldValue struct {
		field string
		value string
	}
	var items []fieldValue
	for k, v := range hash {
		items = append(items, fieldValue{field: k, value: v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].field < items[j].field
	})

	var selectedField string
	var selectedRow int = -1

	table := widget.NewTable(
		func() (int, int) { return len(items), 2 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if id.Col == 0 {
				label.SetText(items[id.Row].field)
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.SetText(items[id.Row].value)
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)
	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 300)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(items) {
			selectedField = items[id.Row].field
			selectedRow = id.Row
			if id.Col == 1 {
				// Click on value column - edit
				ve.showEditValueDialog("Value", items[id.Row].value, func(newVal string) {
					err := ve.client.HashSet(key.Key, selectedField, newVal)
					if err != nil {
						ShowErrorDialog(ve.window, "Error", err)
						return
					}
					ve.LoadKey(key)
				})
				table.UnselectAll()
			}
		}
	}

	fieldEntry := widget.NewEntry()
	fieldEntry.SetPlaceHolder("Field")

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Value")

	setBtn := widget.NewButtonWithIcon("Add/Update", theme.DocumentSaveIcon(), func() {
		if fieldEntry.Text == "" {
			return
		}
		err := ve.client.HashSet(key.Key, fieldEntry.Text, valueEntry.Text)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		fieldEntry.SetText("")
		valueEntry.SetText("")
		ve.LoadKey(key)
	})

	removeBtn := widget.NewButtonWithIcon("Remove Selected", theme.ContentRemoveIcon(), func() {
		if selectedField == "" || selectedRow < 0 {
			return
		}
		err := ve.client.HashDelete(key.Key, selectedField)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		selectedField = ""
		selectedRow = -1
		ve.LoadKey(key)
	})

	hint := widget.NewLabelWithStyle("Click a value to edit inline", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	addBar := container.NewVBox(
		hint,
		container.NewGridWithColumns(2, fieldEntry, valueEntry),
		container.NewHBox(setBtn, removeBtn),
	)

	return container.NewBorder(nil, addBar, nil, nil, table)
}

func (ve *ValueEditor) buildZSetEditor(key models.RedisKey) fyne.CanvasObject {
	members, err := ve.client.GetSortedSet(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	var selectedMember string
	var selectedRow int = -1

	table := widget.NewTable(
		func() (int, int) { return len(members), 2 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if id.Col == 0 {
				label.SetText(fmt.Sprintf("%.4f", members[id.Row].Score))
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.SetText(members[id.Row].Member)
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)
	table.SetColumnWidth(0, 100)
	table.SetColumnWidth(1, 350)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(members) {
			selectedMember = members[id.Row].Member
			selectedRow = id.Row
			if id.Col == 0 {
				// Click on score - edit score
				ve.showEditValueDialog("Score", fmt.Sprintf("%.4f", members[id.Row].Score), func(newVal string) {
					score, err := strconv.ParseFloat(newVal, 64)
					if err != nil {
						ShowErrorDialog(ve.window, "Invalid Score", fmt.Errorf("score must be a valid number: %w", err))
						return
					}
					// Remove and re-add with new score
					if err := ve.client.SortedSetRemove(key.Key, selectedMember); err != nil {
						ShowErrorDialog(ve.window, "Error", err)
						return
					}
					if err := ve.client.SortedSetAdd(key.Key, score, selectedMember); err != nil {
						ShowErrorDialog(ve.window, "Error", err)
						return
					}
					ve.LoadKey(key)
				})
				table.UnselectAll()
			} else if id.Col == 1 {
				// Click on member - edit member
				oldScore := members[id.Row].Score
				ve.showEditValueDialog("Member", selectedMember, func(newVal string) {
					// Remove old and add new
					if err := ve.client.SortedSetRemove(key.Key, selectedMember); err != nil {
						ShowErrorDialog(ve.window, "Error", err)
						return
					}
					if err := ve.client.SortedSetAdd(key.Key, oldScore, newVal); err != nil {
						ShowErrorDialog(ve.window, "Error", err)
						return
					}
					ve.LoadKey(key)
				})
				table.UnselectAll()
			}
		}
	}

	scoreEntry := widget.NewEntry()
	scoreEntry.SetPlaceHolder("Score")

	memberEntry := widget.NewEntry()
	memberEntry.SetPlaceHolder("Member")

	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		if memberEntry.Text == "" {
			return
		}
		score := 0.0
		if scoreEntry.Text != "" {
			var err error
			score, err = strconv.ParseFloat(scoreEntry.Text, 64)
			if err != nil {
				ShowErrorDialog(ve.window, "Invalid Score", fmt.Errorf("score must be a valid number"))
				return
			}
		}
		err := ve.client.SortedSetAdd(key.Key, score, memberEntry.Text)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		scoreEntry.SetText("")
		memberEntry.SetText("")
		ve.LoadKey(key)
	})

	removeBtn := widget.NewButtonWithIcon("Remove Selected", theme.ContentRemoveIcon(), func() {
		if selectedMember == "" || selectedRow < 0 {
			return
		}
		err := ve.client.SortedSetRemove(key.Key, selectedMember)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		selectedMember = ""
		selectedRow = -1
		ve.LoadKey(key)
	})

	hint := widget.NewLabelWithStyle("Click score or member to edit", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	addBar := container.NewVBox(
		hint,
		container.NewGridWithColumns(2, scoreEntry, memberEntry),
		container.NewHBox(addBtn, removeBtn),
	)

	return container.NewBorder(nil, addBar, nil, nil, table)
}

func (ve *ValueEditor) showEditValueDialog(fieldName string, currentValue string, onSave func(string)) {
	entry := widget.NewMultiLineEntry()
	entry.SetText(currentValue)
	entry.Wrapping = fyne.TextWrapWord

	d := dialog.NewForm(fmt.Sprintf("Edit %s", fieldName), "Save", "Cancel",
		[]*widget.FormItem{
			{Text: fieldName, Widget: entry},
		},
		func(save bool) {
			if save {
				onSave(entry.Text)
			}
		}, ve.window)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}

// Clear clears the editor
func (ve *ValueEditor) Clear() {
	ve.currentKey = nil
	ve.keyLabel.SetText("No key selected")
	ve.typeLabel.SetText("")
	ve.ttlLabel.SetText("")
	ve.contentArea.RemoveAll()
	ve.contentArea.Add(widget.NewLabel("Select a key to view its value"))
	ve.contentArea.Refresh()
}
