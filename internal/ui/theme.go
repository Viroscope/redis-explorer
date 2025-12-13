package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"redis-explorer/internal/models"
)

// Common color constants for reuse across themes
var (
	// Standard colors
	colorWhite      = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colorBlack      = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	colorTransparent = color.NRGBA{R: 0, G: 0, B: 0, A: 0}

	// Shadow colors with varying opacity
	colorShadowDark   = color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	colorShadowMedium = color.NRGBA{R: 0, G: 0, B: 0, A: 80}
	colorShadowLight  = color.NRGBA{R: 0, G: 0, B: 0, A: 50}

	// Standard semantic colors
	colorErrorRed     = color.NRGBA{R: 244, G: 67, B: 54, A: 255}
	colorSuccessGreen = color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	colorWarningOrange = color.NRGBA{R: 255, G: 152, B: 0, A: 255}

	// Google Blue (Material Design primary)
	colorGoogleBlue = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	colorMaterialBlue = color.NRGBA{R: 25, G: 118, B: 210, A: 255}
)

// CustomTheme implements fyne.Theme with customizable colors
type CustomTheme struct {
	name            models.ThemeName
	backgroundColor color.Color
	foregroundColor color.Color
	primaryColor    color.Color
	hoverColor      color.Color
	inputBgColor    color.Color
	disabledColor   color.Color
	scrollBarColor  color.Color
	separatorColor  color.Color
	shadowColor     color.Color
	errorColor      color.Color
	successColor    color.Color
	warningColor    color.Color
}

// Dark theme colors
var darkTheme = &CustomTheme{
	name:            models.ThemeDark,
	backgroundColor: color.NRGBA{R: 30, G: 30, B: 30, A: 255},
	foregroundColor: color.NRGBA{R: 230, G: 230, B: 230, A: 255},
	primaryColor:    colorGoogleBlue,
	hoverColor:      color.NRGBA{R: 60, G: 60, B: 60, A: 255},
	inputBgColor:    color.NRGBA{R: 45, G: 45, B: 45, A: 255},
	disabledColor:   color.NRGBA{R: 100, G: 100, B: 100, A: 255},
	scrollBarColor:  color.NRGBA{R: 80, G: 80, B: 80, A: 255},
	separatorColor:  color.NRGBA{R: 60, G: 60, B: 60, A: 255},
	shadowColor:     colorShadowDark,
	errorColor:      colorErrorRed,
	successColor:    colorSuccessGreen,
	warningColor:    colorWarningOrange,
}

// Light theme colors
var lightTheme = &CustomTheme{
	name:            models.ThemeLight,
	backgroundColor: color.NRGBA{R: 250, G: 250, B: 250, A: 255},
	foregroundColor: color.NRGBA{R: 33, G: 33, B: 33, A: 255},
	primaryColor:    colorMaterialBlue,
	hoverColor:      color.NRGBA{R: 230, G: 230, B: 230, A: 255},
	inputBgColor:    colorWhite,
	disabledColor:   color.NRGBA{R: 180, G: 180, B: 180, A: 255},
	scrollBarColor:  color.NRGBA{R: 200, G: 200, B: 200, A: 255},
	separatorColor:  color.NRGBA{R: 220, G: 220, B: 220, A: 255},
	shadowColor:     colorShadowLight,
	errorColor:      color.NRGBA{R: 211, G: 47, B: 47, A: 255},
	successColor:    color.NRGBA{R: 56, G: 142, B: 60, A: 255},
	warningColor:    color.NRGBA{R: 245, G: 124, B: 0, A: 255},
}

// Nord theme colors
var nordTheme = &CustomTheme{
	name:            models.ThemeNord,
	backgroundColor: color.NRGBA{R: 46, G: 52, B: 64, A: 255},
	foregroundColor: color.NRGBA{R: 216, G: 222, B: 233, A: 255},
	primaryColor:    color.NRGBA{R: 136, G: 192, B: 208, A: 255},
	hoverColor:      color.NRGBA{R: 59, G: 66, B: 82, A: 255},
	inputBgColor:    color.NRGBA{R: 59, G: 66, B: 82, A: 255},
	disabledColor:   color.NRGBA{R: 76, G: 86, B: 106, A: 255},
	scrollBarColor:  color.NRGBA{R: 76, G: 86, B: 106, A: 255},
	separatorColor:  color.NRGBA{R: 59, G: 66, B: 82, A: 255},
	shadowColor:     colorShadowMedium,
	errorColor:      color.NRGBA{R: 191, G: 97, B: 106, A: 255},
	successColor:    color.NRGBA{R: 163, G: 190, B: 140, A: 255},
	warningColor:    color.NRGBA{R: 235, G: 203, B: 139, A: 255},
}

// Dracula theme colors
var draculaTheme = &CustomTheme{
	name:            models.ThemeDracula,
	backgroundColor: color.NRGBA{R: 40, G: 42, B: 54, A: 255},
	foregroundColor: color.NRGBA{R: 248, G: 248, B: 242, A: 255},
	primaryColor:    color.NRGBA{R: 189, G: 147, B: 249, A: 255},
	hoverColor:      color.NRGBA{R: 68, G: 71, B: 90, A: 255},
	inputBgColor:    color.NRGBA{R: 68, G: 71, B: 90, A: 255},
	disabledColor:   color.NRGBA{R: 98, G: 114, B: 164, A: 255},
	scrollBarColor:  color.NRGBA{R: 98, G: 114, B: 164, A: 255},
	separatorColor:  color.NRGBA{R: 68, G: 71, B: 90, A: 255},
	shadowColor:     colorShadowDark,
	errorColor:      color.NRGBA{R: 255, G: 85, B: 85, A: 255},
	successColor:    color.NRGBA{R: 80, G: 250, B: 123, A: 255},
	warningColor:    color.NRGBA{R: 255, G: 184, B: 108, A: 255},
}

// Solarized Dark theme colors
var solarizedTheme = &CustomTheme{
	name:            models.ThemeSolarized,
	backgroundColor: color.NRGBA{R: 0, G: 43, B: 54, A: 255},
	foregroundColor: color.NRGBA{R: 131, G: 148, B: 150, A: 255},
	primaryColor:    color.NRGBA{R: 38, G: 139, B: 210, A: 255},
	hoverColor:      color.NRGBA{R: 7, G: 54, B: 66, A: 255},
	inputBgColor:    color.NRGBA{R: 7, G: 54, B: 66, A: 255},
	disabledColor:   color.NRGBA{R: 88, G: 110, B: 117, A: 255},
	scrollBarColor:  color.NRGBA{R: 88, G: 110, B: 117, A: 255},
	separatorColor:  color.NRGBA{R: 7, G: 54, B: 66, A: 255},
	shadowColor:     colorShadowMedium,
	errorColor:      color.NRGBA{R: 220, G: 50, B: 47, A: 255},
	successColor:    color.NRGBA{R: 133, G: 153, B: 0, A: 255},
	warningColor:    color.NRGBA{R: 203, G: 75, B: 22, A: 255},
}

// GetTheme returns the theme for the given name
func GetTheme(name models.ThemeName) fyne.Theme {
	switch name {
	case models.ThemeLight:
		return lightTheme
	case models.ThemeNord:
		return nordTheme
	case models.ThemeDracula:
		return draculaTheme
	case models.ThemeSolarized:
		return solarizedTheme
	default:
		return darkTheme
	}
}

// Color implements fyne.Theme
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.backgroundColor
	case theme.ColorNameForeground:
		return t.foregroundColor
	case theme.ColorNamePrimary:
		return t.primaryColor
	case theme.ColorNameHover:
		return t.hoverColor
	case theme.ColorNameInputBackground:
		return t.inputBgColor
	case theme.ColorNameDisabled:
		return t.disabledColor
	case theme.ColorNameScrollBar:
		return t.scrollBarColor
	case theme.ColorNameSeparator:
		return t.separatorColor
	case theme.ColorNameShadow:
		return t.shadowColor
	case theme.ColorNameError:
		return t.errorColor
	case theme.ColorNameSuccess:
		return t.successColor
	case theme.ColorNameWarning:
		return t.warningColor
	case theme.ColorNameButton:
		return t.inputBgColor
	case theme.ColorNameDisabledButton:
		return t.disabledColor
	case theme.ColorNamePlaceHolder:
		return t.disabledColor
	case theme.ColorNamePressed:
		return t.primaryColor
	case theme.ColorNameFocus:
		return t.primaryColor
	case theme.ColorNameSelection:
		return color.NRGBA{
			R: t.primaryColor.(color.NRGBA).R,
			G: t.primaryColor.(color.NRGBA).G,
			B: t.primaryColor.(color.NRGBA).B,
			A: 100,
		}
	case theme.ColorNameHeaderBackground:
		return t.hoverColor
	case theme.ColorNameInputBorder:
		return t.separatorColor
	case theme.ColorNameMenuBackground:
		return t.backgroundColor
	case theme.ColorNameOverlayBackground:
		return t.backgroundColor
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font implements fyne.Theme
func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon implements fyne.Theme
func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size implements fyne.Theme
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 1
	}
	return theme.DefaultTheme().Size(name)
}
