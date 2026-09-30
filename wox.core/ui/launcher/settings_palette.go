package launcher

// KobeTools fork: ⌘K / Ctrl+K command palette for the Settings window. It is a
// second front end over the rail search: results come from settingsSearchResults
// and activation goes through activateSettingsSearchResult, so ranking, navigation,
// and the destination highlight stay identical. Kept in its own file so upstream
// syncs only touch the small hook-ins in app.go, windows.go, and settings_adapter.go.

import (
	"runtime"
	"strings"

	launcherview "wox/ui/launcher/view"
	woxui "wox/ui/runtime"
	woxwidget "wox/ui/widget"
)

// settingsCommandPaletteState is the palette's transient interaction state. The zero value is closed.
type settingsCommandPaletteState struct {
	open     bool
	editor   *woxwidget.TextEditingController
	query    string
	selected int
	// restoreRailSearch records that the rail search field owned focus when the
	// palette opened. The rail field is blurred on open so the Host has nothing to
	// restore after an activation (matching rail-search navigation, which clears
	// focus); dismissing the palette re-focuses it explicitly instead.
	restoreRailSearch bool
}

func (p *settingsCommandPaletteState) Open() bool { return p.open }

// Query returns the trimmed palette query used for ranking.
func (p *settingsCommandPaletteState) Query() string { return strings.TrimSpace(p.query) }

func (p *settingsCommandPaletteState) Selected() int { return p.selected }

// show opens the palette with an empty query and the first row selected.
func (p *settingsCommandPaletteState) show(restoreRailSearch bool) {
	*p = settingsCommandPaletteState{open: true, editor: woxwidget.NewTextEditingController(""), restoreRailSearch: restoreRailSearch}
}

// hide closes the palette and reports whether the rail search should get focus back.
func (p *settingsCommandPaletteState) hide() (restoreRailSearch bool) {
	restoreRailSearch = p.open && p.restoreRailSearch
	*p = settingsCommandPaletteState{}
	return restoreRailSearch
}

// setQuery follows the field text and resets the selection to the best match.
func (p *settingsCommandPaletteState) setQuery(value string) {
	if p.editor != nil && p.editor.Text() != value {
		p.editor.SetText(value, false)
	}
	p.query = value
	p.selected = 0
}

// move shifts the selection and clamps it to the current result count.
func (p *settingsCommandPaletteState) move(delta, count int) {
	if count <= 0 {
		p.selected = 0
		return
	}
	p.selected = min(max(0, p.selected+delta), count-1)
}

// selectedIndex clamps the stored selection to count, returning -1 when there is nothing to open.
func (p *settingsCommandPaletteState) selectedIndex(count int) int {
	if count <= 0 {
		return -1
	}
	return min(max(0, p.selected), count-1)
}

// isSettingsPaletteShortcut matches exactly ⌘K on macOS and Ctrl+K elsewhere.
func isSettingsPaletteShortcut(event woxui.KeyEvent) bool {
	primary := woxui.KeyModifierControl
	if runtime.GOOS == "darwin" {
		primary = woxui.KeyModifierMeta
	}
	return event.Key == woxui.Key("k") && event.Modifiers == primary
}

// settingsPaletteShortcutLabel is the platform label shown in the rail search hint.
func settingsPaletteShortcutLabel() string {
	if runtime.GOOS == "darwin" {
		return "⌘K"
	}
	return "Ctrl+K"
}

// onSettingsPaletteKey runs before focused widgets so the shortcut works from any control,
// and owns navigation keys while the palette is open. Printable keys fall through to the field.
func (a *App) onSettingsPaletteKey(event woxui.KeyEvent) bool {
	if !event.Down || event.Composing {
		return false
	}
	shortcut := isSettingsPaletteShortcut(event)
	if !a.settingsCommandPalette.Open() {
		if !shortcut || event.Repeat {
			return false
		}
		return a.openSettingsPalette()
	}
	switch {
	case shortcut:
		if !event.Repeat {
			a.closeSettingsPalette()
		}
	case event.Key == woxui.KeyEscape:
		a.closeSettingsPalette()
	case event.Key == woxui.KeyArrowDown:
		a.moveSettingsPaletteSelection(1)
	case event.Key == woxui.KeyArrowUp:
		a.moveSettingsPaletteSelection(-1)
	case event.Key == woxui.KeyEnter:
		a.activateSelectedSettingsPaletteResult()
	default:
		return false
	}
	return true
}

// settingsPaletteBlocked mirrors the modal overlays in buildSettings; the palette never stacks on them.
func (a *App) settingsPaletteBlocked(snapshot settingsSnapshot) bool {
	return snapshot.tableEditor != nil || snapshot.ai.ModelManager != nil || snapshot.general.ShowDisplayPicker != nil ||
		snapshot.general.ChoicePicker != nil || snapshot.cloud.PluginDialog != nil || snapshot.cloud.Form != nil ||
		snapshot.privacy.Sample != "" || (snapshot.tab == "theme" && snapshot.theme.AutoEditor != nil) || a.settingsDemo != nil
}

// openSettingsPalette shows the palette unless another settings dialog already owns the window.
func (a *App) openSettingsPalette() bool {
	if a.settingsPaletteBlocked(a.settingsSnapshot()) {
		return false
	}
	railSearchFocused := a.settingsSearch.Focused()
	if railSearchFocused {
		a.blurSettingsSearch()
	}
	a.settingsCommandPalette.show(railSearchFocused)
	a.invalidateSettingsWindow()
	return true
}

// closeSettingsPalette dismisses without navigating. The Host restores any other prior focus
// when the modal scope unmounts; the rail search is re-armed through its Autofocus state.
func (a *App) closeSettingsPalette() {
	if a.settingsCommandPalette.hide() {
		a.setSettingsSearchFocused(true)
	}
	a.invalidateSettingsWindow()
}

func (a *App) setSettingsPaletteQuery(value string) error {
	a.settingsCommandPalette.setQuery(value)
	a.invalidateSettingsWindow()
	return nil
}

// settingsPaletteResults ranks the palette query with the rail search index.
func (a *App) settingsPaletteResults(snapshot settingsSnapshot) []settingsSearchResult {
	snapshot.search.Query = woxui.TextEditingState{Text: a.settingsCommandPalette.Query()}
	return a.settingsSearchResults(snapshot)
}

func (a *App) moveSettingsPaletteSelection(delta int) {
	a.settingsCommandPalette.move(delta, len(a.settingsPaletteResults(a.settingsSnapshot())))
	a.invalidateSettingsWindow()
}

func (a *App) selectSettingsPaletteResult(index int) {
	if index == a.settingsCommandPalette.Selected() {
		return
	}
	a.settingsCommandPalette.selected = index
	a.invalidateSettingsWindow()
}

func (a *App) activateSelectedSettingsPaletteResult() {
	results := a.settingsPaletteResults(a.settingsSnapshot())
	if index := a.settingsCommandPalette.selectedIndex(len(results)); index >= 0 {
		a.activateSettingsPaletteResult(results[index])
	}
}

// activateSettingsPaletteResult closes the palette without restoring focus, then reuses
// the rail search navigation so tab routing, plugin loading, and the highlight match it.
func (a *App) activateSettingsPaletteResult(result settingsSearchResult) {
	a.settingsCommandPalette.hide()
	a.activateSettingsSearchResult(result)
	a.invalidateSettingsWindow()
}

// buildSettingsPaletteOverlay prepares the palette rows with the same icons and labels as the rail results.
func (a *App) buildSettingsPaletteOverlay(snapshot settingsSnapshot, width, height, imageScale float32) woxwidget.Widget {
	query := a.settingsCommandPalette.Query()
	var results []settingsSearchResult
	if query != "" {
		results = a.settingsPaletteResults(snapshot)
	}
	selected := a.settingsCommandPalette.selectedIndex(len(results))
	items := make([]launcherview.SettingsSearchResult, 0, len(results))
	for index, result := range results {
		iconTint := snapshot.palette.TextSecondary
		if index == selected {
			iconTint = snapshot.palette.SelectionText
		}
		icon := a.imageForTint(settingsSearchResultIconSource(result.kind), &iconTint, physicalImageSize(24, imageScale))
		if (result.kind == settingsSearchPlugin || result.kind == settingsSearchPluginSetting) && result.icon.ImageData != "" {
			if pluginIcon := a.imageForSurface(result.icon, physicalImageSize(24, imageScale), settingsPalette().Background); pluginIcon != nil {
				icon = pluginIcon
			}
		}
		items = append(items, launcherview.SettingsSearchResult{
			Title: result.title, Subtitle: a.settingsSearchResultTypeLabel(result.kind) + " · " + result.subtitle, Icon: icon,
			OnHover: func() { a.selectSettingsPaletteResult(index) }, OnTap: func() { a.activateSettingsPaletteResult(result) },
		})
	}
	emptyMessage := a.translate("i18n:ui_setting_search_empty")
	if len(results) == 0 {
		if snapshot.search.Loading {
			emptyMessage = a.translate("i18n:ui_cloud_sync_plugin_exclusions_loading")
		} else if snapshot.search.Error != "" {
			emptyMessage = snapshot.search.Error
		}
	}
	iconTint := snapshot.palette.TextSecondary
	return launcherview.SettingsCommandPalette(launcherview.SettingsCommandPaletteProps{
		Width: width, Height: height, Label: a.translate("i18n:ui_setting_search_placeholder"), DismissLabel: a.translate("i18n:ui_close"), Placeholder: a.translate("i18n:ui_setting_palette_placeholder"),
		Query: query, Controller: a.settingsCommandPalette.editor, Window: a.settingsNativeWindow(), Theme: snapshot.palette,
		SearchIcon: a.imageForTint(settingControlIconSource("search"), &iconTint, physicalImageSize(18, imageScale)),
		Results:    items, Selected: selected, EmptyMessage: emptyMessage, Footer: a.translate("i18n:ui_setting_palette_footer"),
		OnChanged: func(value string) { _ = a.setSettingsPaletteQuery(value) }, OnSetValue: a.setSettingsPaletteQuery,
		OnClear: func() { _ = a.setSettingsPaletteQuery("") }, OnDismiss: a.closeSettingsPalette,
	})
}
