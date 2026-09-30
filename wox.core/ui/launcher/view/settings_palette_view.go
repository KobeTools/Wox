package view

// KobeTools fork: the ⌘K / Ctrl+K settings command palette. It lives in its own
// file so upstream syncs of the rail search views stay conflict-free.

import (
	"fmt"

	woxcomponent "wox/ui/launcher/component"
	woxui "wox/ui/runtime"
	woxwidget "wox/ui/widget"
)

const (
	SettingsCommandPaletteMaxWidth = float32(560)
	// SettingsCommandPaletteFieldKey is the palette's search field; the dialog focuses it on mount.
	SettingsCommandPaletteFieldKey = "settings-palette-field"

	settingsPalettePadding   = float32(12)
	settingsPaletteGap       = float32(8)
	settingsPaletteRowHeight = float32(52)
	settingsPaletteFooter    = float32(18)
	settingsPaletteMargin    = float32(24)
)

// SettingsCommandPaletteProps contains the prepared palette query, ranked rows, and actions.
type SettingsCommandPaletteProps struct {
	// Width and Height are the full settings window; the scrim covers all of it.
	Width        float32
	Height       float32
	Label        string
	DismissLabel string
	Placeholder  string
	Query        string
	Controller   *woxwidget.TextEditingController
	SearchIcon   *woxui.Image
	Window       *woxui.Window
	Results      []SettingsSearchResult
	Selected     int
	EmptyMessage string
	Footer       string
	Theme        woxcomponent.ControlTheme
	OnChanged    func(string)
	OnSetValue   func(string) error
	OnClear      func()
	OnDismiss    func()
}

// SettingsCommandPalettePanelFrame places the panel horizontally centered in the upper part of the window.
// The top edge stays fixed while results change, so the field does not jump as the user types.
func SettingsCommandPalettePanelFrame(width, height float32) (left, top, panelWidth float32) {
	panelWidth = max(float32(0), min(SettingsCommandPaletteMaxWidth, width-2*settingsPaletteMargin))
	left = max(float32(0), (width-panelWidth)/2)
	top = max(settingsPaletteMargin, height*0.14)
	return left, top, panelWidth
}

// SettingsCommandPalette builds the dimmed scrim and the modal search panel above it.
func SettingsCommandPalette(props SettingsCommandPaletteProps) woxwidget.Widget {
	left, top, panelWidth := SettingsCommandPalettePanelFrame(props.Width, props.Height)
	innerWidth := max(float32(0), panelWidth-2*settingsPalettePadding)
	selected := 0
	if len(props.Results) > 0 {
		selected = min(max(0, props.Selected), len(props.Results)-1)
	}

	field := woxcomponent.WoxSearchField(woxcomponent.SearchFieldProps{
		ID: SettingsCommandPaletteFieldKey, Label: props.Placeholder, Width: innerWidth, Value: props.Query, Focused: true, Autofocus: true,
		Controller: props.Controller, SearchIcon: props.SearchIcon, Window: props.Window, Theme: props.Theme,
		OnClear: props.OnClear, OnChanged: props.OnChanged, OnSetValue: props.OnSetValue,
	})
	children := []woxwidget.Widget{field}
	contentHeight := woxcomponent.SettingsSearchHeight

	if props.Query != "" {
		// The list may shrink to keep the panel inside the window; it scrolls beyond that.
		fixed := 2*settingsPalettePadding + woxcomponent.SettingsSearchHeight + 2*settingsPaletteGap + settingsPaletteFooter
		maxList := max(settingsPaletteRowHeight, props.Height-top-settingsPaletteMargin-fixed)
		var list woxwidget.Widget
		listHeight := settingsPaletteRowHeight
		if len(props.Results) == 0 {
			list = woxwidget.Container{Width: innerWidth, Height: listHeight, Padding: woxwidget.Insets{Left: 10}, Child: woxwidget.Align{Height: listHeight, Vertical: 0.5, Child: woxwidget.Text{
				Value: props.EmptyMessage, Style: woxui.TextStyle{Size: props.Theme.Scaled(woxcomponent.SettingsSearchTitleFontSize)}, Color: props.Theme.TextSecondary,
			}}}
		} else {
			listHeight = min(maxList, float32(len(props.Results))*settingsPaletteRowHeight)
			list = settingsPaletteRows(props, selected, innerWidth, listHeight)
		}
		children = append(children, list)
		contentHeight += settingsPaletteGap + listHeight
	}

	children = append(children, woxwidget.Container{Width: innerWidth, Height: settingsPaletteFooter, Padding: woxwidget.Insets{Left: 4}, Child: woxwidget.Align{Height: settingsPaletteFooter, Vertical: 0.5, Child: woxwidget.Text{
		Value: props.Footer, Style: woxui.TextStyle{Size: props.Theme.Scaled(woxcomponent.SettingsSearchSubtitleFontSize)}, Color: props.Theme.TextSecondary,
	}}})
	contentHeight += settingsPaletteGap + settingsPaletteFooter

	panel := woxcomponent.WoxDialog(woxcomponent.DialogProps{
		ID: "settings-palette", Label: props.Label, Width: panelWidth, Height: contentHeight + 2*settingsPalettePadding,
		Solid: true, Radius: 12, Padding: woxwidget.UniformInsets(settingsPalettePadding), InitialFocus: SettingsCommandPaletteFieldKey,
		OnEscape: props.OnDismiss, Theme: props.Theme,
		Child: woxwidget.Flex{Axis: woxwidget.Vertical, Gap: settingsPaletteGap, Children: children},
	})
	scrim := woxwidget.Semantics{
		AutomationID: "settings-palette-scrim", Role: woxui.AccessibilityRoleButton, Label: props.DismissLabel,
		Child: woxwidget.Gesture{ID: "settings-palette-scrim", OnTap: props.OnDismiss, OnScroll: func(woxui.Point) {}, Child: woxwidget.Container{
			Width: props.Width, Height: props.Height, Color: woxui.Color{A: 112},
		}},
	}
	return woxwidget.Stack{Width: props.Width, Height: props.Height, Children: []woxwidget.StackChild{
		{Child: scrim},
		{Left: left, Top: top, Child: panel},
	}}
}

// settingsPaletteRows renders ranked results with the same row content as the rail search panel.
func settingsPaletteRows(props SettingsCommandPaletteProps, selected int, width, height float32) woxwidget.Widget {
	rows := make([]woxwidget.Widget, 0, len(props.Results))
	for index, result := range props.Results {
		background := woxui.Color{}
		titleColor := props.Theme.Text
		subtitleColor := props.Theme.TextSecondary
		if index == selected {
			background = props.Theme.SelectionBackground
			titleColor = props.Theme.SelectionText
			subtitleColor = props.Theme.SelectionText
		}
		text := woxwidget.Flex{Axis: woxwidget.Vertical, Gap: 3, Children: []woxwidget.Widget{
			woxwidget.Text{Value: result.Title, Style: woxui.TextStyle{Size: props.Theme.Scaled(woxcomponent.SettingsSearchTitleFontSize), Weight: woxui.FontWeightSemibold}, Color: titleColor},
			woxwidget.Text{Value: result.Subtitle, Style: woxui.TextStyle{Size: props.Theme.Scaled(woxcomponent.SettingsSearchSubtitleFontSize)}, Color: subtitleColor},
		}}
		var content woxwidget.Widget = text
		if result.Icon != nil {
			content = woxwidget.Flex{Axis: woxwidget.Horizontal, Gap: 10, CrossAxisAlignment: woxwidget.CrossAxisCenter, Children: []woxwidget.Widget{
				woxwidget.Align{Width: 24, Height: 38, Horizontal: 0.5, Vertical: 0.5, Child: woxwidget.Image{Source: result.Icon, Width: 24, Height: 24}},
				woxwidget.Expanded{Child: woxwidget.Align{Height: 38, Vertical: 0.5, Child: text}},
			}}
		}
		onHover := result.OnHover
		rows = append(rows, woxwidget.Gesture{ID: fmt.Sprintf("settings-palette-result-%d", index), OnHover: func(inside bool) {
			if inside && onHover != nil {
				onHover()
			}
		}, OnTap: result.OnTap, Child: woxwidget.Container{
			Width: width, Height: settingsPaletteRowHeight, Radius: 6, Color: background, Padding: woxwidget.Insets{Left: 10, Right: 10},
			Child: woxwidget.Align{Height: settingsPaletteRowHeight, Vertical: 0.5, Child: content},
		}})
	}
	start := float32(selected) * settingsPaletteRowHeight
	return woxcomponent.WoxScrollView(woxcomponent.ScrollViewProps{
		Key: "settings-palette-results", Width: width, Height: height,
		KeepVisible: &woxwidget.ScrollRange{Start: start, End: start + settingsPaletteRowHeight},
		Content:     woxwidget.Flex{Axis: woxwidget.Vertical, Children: rows}, Theme: props.Theme, ThumbColor: props.Theme.Text,
	})
}
