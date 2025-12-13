package ui

import (
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"redis-explorer/internal/config"
	"redis-explorer/internal/models"
)

const (
	AppVersion = "1.0.0"
	AppName    = "Redis Explorer"
)

// ShowConnectionDialog shows a dialog to add or edit a connection
func ShowConnectionDialog(window fyne.Window, conn *models.ServerConnection, onSave func(models.ServerConnection)) {
	isNew := conn == nil
	if isNew {
		conn = &models.ServerConnection{
			ID:       uuid.New().String(),
			Host:     "localhost",
			Port:     6379,
			Database: 0,
		}
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(conn.Name)
	nameEntry.SetPlaceHolder("Connection name")

	hostEntry := widget.NewEntry()
	hostEntry.SetText(conn.Host)
	hostEntry.SetPlaceHolder("localhost")

	portEntry := widget.NewEntry()
	portEntry.SetText(strconv.Itoa(conn.Port))
	portEntry.SetPlaceHolder("6379")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(conn.Password)
	passwordEntry.SetPlaceHolder("Optional")

	dbEntry := widget.NewEntry()
	dbEntry.SetText(strconv.Itoa(conn.Database))
	dbEntry.SetPlaceHolder("0")

	tlsCheck := widget.NewCheck("Use TLS", nil)
	tlsCheck.SetChecked(conn.UseTLS)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Host", Widget: hostEntry},
			{Text: "Port", Widget: portEntry},
			{Text: "Password", Widget: passwordEntry},
			{Text: "Database", Widget: dbEntry},
			{Text: "", Widget: tlsCheck},
		},
	}

	title := "Add Connection"
	if !isNew {
		title = "Edit Connection"
	}

	d := dialog.NewCustomConfirm(title, "Save", "Cancel", form, func(save bool) {
		if !save {
			return
		}

		port, err := strconv.Atoi(portEntry.Text)
		if err != nil {
			port = 6379
		}

		db, err := strconv.Atoi(dbEntry.Text)
		if err != nil {
			db = 0
		}

		newConn := models.ServerConnection{
			ID:       conn.ID,
			Name:     nameEntry.Text,
			Host:     hostEntry.Text,
			Port:     port,
			Password: passwordEntry.Text,
			Database: db,
			UseTLS:   tlsCheck.Checked,
		}

		if newConn.Name == "" {
			newConn.Name = newConn.Host
		}

		onSave(newConn)
	}, window)

	d.Resize(fyne.NewSize(400, 300))
	d.Show()
}

// ShowThemeDialog shows a dialog to select the theme
func ShowThemeDialog(window fyne.Window, currentTheme models.ThemeName, onSelect func(models.ThemeName)) {
	themes := models.AllThemes()
	var options []string
	selectedIndex := 0

	for i, t := range themes {
		options = append(options, t.DisplayName())
		if t == currentTheme {
			selectedIndex = i
		}
	}

	selector := widget.NewSelect(options, nil)
	selector.SetSelectedIndex(selectedIndex)

	d := dialog.NewCustomConfirm("Select Theme", "Apply", "Cancel",
		container.NewVBox(
			widget.NewLabel("Choose your preferred theme:"),
			selector,
		),
		func(apply bool) {
			if apply && selector.SelectedIndex() >= 0 {
				onSelect(themes[selector.SelectedIndex()])
			}
		}, window)

	d.Resize(fyne.NewSize(300, 150))
	d.Show()
}

// ShowConfirmDialog shows a confirmation dialog
func ShowConfirmDialog(window fyne.Window, title, message string, onConfirm func()) {
	dialog.ShowConfirm(title, message, func(confirmed bool) {
		if confirmed {
			onConfirm()
		}
	}, window)
}

// ShowErrorDialog shows an error dialog
func ShowErrorDialog(window fyne.Window, title string, err error) {
	dialog.ShowError(err, window)
}

// ShowInfoDialog shows an information dialog
func ShowInfoDialog(window fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, window)
}

// ShowNewKeyDialog shows a dialog to create a new key
func ShowNewKeyDialog(window fyne.Window, onCreate func(key string, keyType string)) {
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Key name")

	typeSelect := widget.NewSelect([]string{"string", "list", "set", "hash", "zset"}, nil)
	typeSelect.SetSelected("string")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Key", Widget: keyEntry},
			{Text: "Type", Widget: typeSelect},
		},
	}

	d := dialog.NewCustomConfirm("New Key", "Create", "Cancel", form, func(create bool) {
		if create && keyEntry.Text != "" {
			onCreate(keyEntry.Text, typeSelect.Selected)
		}
	}, window)

	d.Resize(fyne.NewSize(350, 150))
	d.Show()
}

// ShowTTLDialog shows a dialog to set TTL
func ShowTTLDialog(window fyne.Window, currentTTL int64, onSet func(ttl int64)) {
	ttlEntry := widget.NewEntry()
	if currentTTL > 0 {
		ttlEntry.SetText(strconv.FormatInt(currentTTL, 10))
	}
	ttlEntry.SetPlaceHolder("Seconds (0 or empty for no expiry)")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "TTL (seconds)", Widget: ttlEntry},
		},
	}

	d := dialog.NewCustomConfirm("Set TTL", "Set", "Cancel", form, func(set bool) {
		if set {
			ttl, err := strconv.ParseInt(ttlEntry.Text, 10, 64)
			if err != nil {
				ttl = 0
			}
			onSet(ttl)
		}
	}, window)

	d.Resize(fyne.NewSize(300, 120))
	d.Show()
}

// ShowSettingsDialog shows the settings dialog
func ShowSettingsDialog(window fyne.Window, onSave func()) {
	cfg := config.Get()

	scanCountEntry := widget.NewEntry()
	scanCountEntry.SetText(strconv.Itoa(cfg.KeyScanCount))

	refreshEntry := widget.NewEntry()
	refreshEntry.SetText(strconv.Itoa(cfg.AutoRefreshSecs))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Key Scan Count", Widget: scanCountEntry, HintText: "Number of keys to scan per request"},
			{Text: "Auto Refresh (sec)", Widget: refreshEntry, HintText: "0 to disable"},
		},
	}

	d := dialog.NewCustomConfirm("Settings", "Save", "Cancel", form, func(save bool) {
		if save {
			scanCount, err := strconv.Atoi(scanCountEntry.Text)
			if err == nil && scanCount > 0 {
				cfg.KeyScanCount = scanCount
			}

			refresh, err := strconv.Atoi(refreshEntry.Text)
			if err == nil && refresh >= 0 {
				cfg.AutoRefreshSecs = refresh
			}

			config.Save()
			if onSave != nil {
				onSave()
			}
		}
	}, window)

	d.Resize(fyne.NewSize(400, 180))
	d.Show()
}

// ShowAboutDialog shows a professional about dialog
func ShowAboutDialog(window fyne.Window, icon fyne.Resource) {
	// Logo
	var logoImage *canvas.Image
	if icon != nil {
		logoImage = canvas.NewImageFromResource(icon)
		logoImage.SetMinSize(fyne.NewSize(100, 100))
		logoImage.FillMode = canvas.ImageFillContain
	}

	// App info
	titleLabel := widget.NewLabelWithStyle(AppName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	versionLabel := widget.NewLabelWithStyle("Version "+AppVersion, fyne.TextAlignCenter, fyne.TextStyle{})
	descLabel := widget.NewLabelWithStyle(
		"A powerful GUI client for Redis databases.\nSupports all Redis data types with intuitive editing.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Separator
	sep1 := widget.NewSeparator()

	// Developer info
	devHeader := widget.NewLabelWithStyle("Developer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	devName := widget.NewLabelWithStyle("Dark Angel", fyne.TextAlignCenter, fyne.TextStyle{})

	// Discord info
	sep2 := widget.NewSeparator()
	discordHeader := widget.NewLabelWithStyle("Community", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	discordURL, _ := url.Parse("https://discord.gg/swmy25fFHY")
	discordLink := widget.NewHyperlink("Join Arcturus on Discord", discordURL)
	discordLink.Alignment = fyne.TextAlignCenter

	discordInfo := widget.NewLabelWithStyle(
		"Server: Arcturus\nUser ID: 490662159508832287\nServer ID: 1122592718544179251",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	// Tech info
	sep3 := widget.NewSeparator()
	techLabel := widget.NewLabelWithStyle(
		"Built with Go & Fyne",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Layout
	var content *fyne.Container
	if logoImage != nil {
		content = container.NewVBox(
			container.NewCenter(logoImage),
			titleLabel,
			versionLabel,
			descLabel,
			sep1,
			devHeader,
			devName,
			sep2,
			discordHeader,
			container.NewCenter(discordLink),
			discordInfo,
			sep3,
			techLabel,
		)
	} else {
		content = container.NewVBox(
			titleLabel,
			versionLabel,
			descLabel,
			sep1,
			devHeader,
			devName,
			sep2,
			discordHeader,
			container.NewCenter(discordLink),
			discordInfo,
			sep3,
			techLabel,
		)
	}

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(350, 400))

	d := dialog.NewCustom("About "+AppName, "Close", scroll, window)
	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}
