package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"redis-explorer/internal/models"
	"redis-explorer/internal/redis"
)

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

	return container.NewBorder(nil, saveBtn, nil, nil, entry)
}

func (ve *ValueEditor) buildListEditor(key models.RedisKey) fyne.CanvasObject {
	items, err := ve.client.GetList(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("0"),
				widget.NewLabel("value"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(fmt.Sprintf("[%d]", i))
			box.Objects[1].(*widget.Label).SetText(items[i])
		},
	)

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

	addBar := container.NewBorder(nil, nil, nil,
		container.NewHBox(addLeftBtn, addRightBtn),
		addEntry,
	)

	return container.NewBorder(nil, addBar, nil, nil, list)
}

func (ve *ValueEditor) buildSetEditor(key models.RedisKey) fyne.CanvasObject {
	members, err := ve.client.GetSet(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	var selectedMember string

	list := widget.NewList(
		func() int { return len(members) },
		func() fyne.CanvasObject {
			return widget.NewLabel("member")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(members[i])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(members) {
			selectedMember = members[id]
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

	removeBtn := widget.NewButtonWithIcon("Remove", theme.ContentRemoveIcon(), func() {
		if selectedMember == "" {
			return
		}
		err := ve.client.SetRemove(key.Key, selectedMember)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		ve.LoadKey(key)
	})

	addBar := container.NewBorder(nil, nil, nil,
		container.NewHBox(addBtn, removeBtn),
		addEntry,
	)

	return container.NewBorder(nil, addBar, nil, nil, list)
}

func (ve *ValueEditor) buildHashEditor(key models.RedisKey) fyne.CanvasObject {
	hash, err := ve.client.GetHash(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	// Convert map to slice for list
	type fieldValue struct {
		field string
		value string
	}
	var items []fieldValue
	for k, v := range hash {
		items = append(items, fieldValue{field: k, value: v})
	}

	var selectedField string

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("field"),
				widget.NewLabel(":"),
				widget.NewLabel("value"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(items[i].field)
			box.Objects[2].(*widget.Label).SetText(items[i].value)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(items) {
			selectedField = items[id].field
		}
	}

	fieldEntry := widget.NewEntry()
	fieldEntry.SetPlaceHolder("Field")

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Value")

	setBtn := widget.NewButtonWithIcon("Set", theme.DocumentSaveIcon(), func() {
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

	removeBtn := widget.NewButtonWithIcon("Remove", theme.ContentRemoveIcon(), func() {
		if selectedField == "" {
			return
		}
		err := ve.client.HashDelete(key.Key, selectedField)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		ve.LoadKey(key)
	})

	addBar := container.NewVBox(
		container.NewGridWithColumns(2, fieldEntry, valueEntry),
		container.NewHBox(setBtn, removeBtn),
	)

	return container.NewBorder(nil, addBar, nil, nil, list)
}

func (ve *ValueEditor) buildZSetEditor(key models.RedisKey) fyne.CanvasObject {
	members, err := ve.client.GetSortedSet(key.Key)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	var selectedMember string

	list := widget.NewList(
		func() int { return len(members) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("score"),
				widget.NewLabel(":"),
				widget.NewLabel("member"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%.2f", members[i].Score))
			box.Objects[2].(*widget.Label).SetText(members[i].Member)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(members) {
			selectedMember = members[id].Member
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
		score, err := strconv.ParseFloat(scoreEntry.Text, 64)
		if err != nil {
			score = 0
		}
		err = ve.client.SortedSetAdd(key.Key, score, memberEntry.Text)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		scoreEntry.SetText("")
		memberEntry.SetText("")
		ve.LoadKey(key)
	})

	removeBtn := widget.NewButtonWithIcon("Remove", theme.ContentRemoveIcon(), func() {
		if selectedMember == "" {
			return
		}
		err := ve.client.SortedSetRemove(key.Key, selectedMember)
		if err != nil {
			ShowErrorDialog(ve.window, "Error", err)
			return
		}
		ve.LoadKey(key)
	})

	addBar := container.NewVBox(
		container.NewGridWithColumns(2, scoreEntry, memberEntry),
		container.NewHBox(addBtn, removeBtn),
	)

	return container.NewBorder(nil, addBar, nil, nil, list)
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
