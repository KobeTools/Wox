package launcher

// KobeTools fork: ⌘K / Ctrl+K settings command palette.

import (
	"reflect"
	"runtime"
	"testing"

	woxui "wox/ui/runtime"
	woxwidget "wox/ui/widget"
)

func settingsPalettePrimaryModifier() woxui.KeyModifiers {
	if runtime.GOOS == "darwin" {
		return woxui.KeyModifierMeta
	}
	return woxui.KeyModifierControl
}

func newSettingsPaletteTestApp(t *testing.T) *App {
	t.Helper()
	app := newApp(false, nil, woxui.NewWindowManager(), newAppInstanceRegistry(), nil, true, "", launcherWindowID)
	t.Cleanup(app.cancel)
	t.Cleanup(app.clearSettingsSearchHighlight)
	app.settingsOpen = true
	app.settingTab = "general"
	app.settingsSearch.SetEditor(woxwidget.NewTextEditingController(""))
	return app
}

func TestSettingsPaletteShortcutTogglesOpenAndClosed(t *testing.T) {
	app := newSettingsPaletteTestApp(t)
	shortcut := woxui.KeyEvent{Key: "k", Modifiers: settingsPalettePrimaryModifier(), Down: true}

	for _, ignored := range []woxui.KeyEvent{
		{Key: "k", Modifiers: settingsPalettePrimaryModifier()},
		{Key: "k", Modifiers: settingsPalettePrimaryModifier() | woxui.KeyModifierShift, Down: true},
		{Key: "k", Down: true},
		{Key: "k", Modifiers: settingsPalettePrimaryModifier(), Down: true, Composing: true},
	} {
		if app.onSettingsPaletteKey(ignored) || app.settingsCommandPalette.Open() {
			t.Fatalf("event %+v opened the palette", ignored)
		}
	}
	if !app.onSettingsPaletteKey(shortcut) || !app.settingsCommandPalette.Open() {
		t.Fatal("primary+K did not open the palette")
	}
	repeat := shortcut
	repeat.Repeat = true
	if !app.onSettingsPaletteKey(repeat) || !app.settingsCommandPalette.Open() {
		t.Fatal("a held primary+K should be consumed without toggling the palette closed")
	}
	if !app.onSettingsPaletteKey(shortcut) || app.settingsCommandPalette.Open() {
		t.Fatal("primary+K again did not close the palette")
	}
}

func TestSettingsPaletteEscapeClosesAndRestoresRailSearch(t *testing.T) {
	app := newSettingsPaletteTestApp(t)
	app.settingsSearch.SetFocused(true)

	if !app.openSettingsPalette() {
		t.Fatal("palette did not open")
	}
	if app.settingsSearch.Focused() {
		t.Fatal("rail search kept focus under the palette")
	}
	_ = app.setSettingsPaletteQuery("general")
	if app.onSettingsPaletteKey(woxui.KeyEvent{Key: "a", Down: true}) {
		t.Fatal("printable keys must reach the palette field")
	}
	row := app.settingRow
	if app.onSettingsKey(woxui.KeyEvent{Key: woxui.KeySpace, Down: true}) || app.onSettingsKey(woxui.KeyEvent{Key: woxui.KeyArrowDown, Down: true}) || app.settingRow != row {
		t.Fatal("the settings page behind the open palette handled a key")
	}
	if !app.onSettingsPaletteKey(woxui.KeyEvent{Key: woxui.KeyEscape, Down: true}) || app.settingsCommandPalette.Open() {
		t.Fatal("Escape did not close the palette")
	}
	if !app.settingsSearch.Focused() {
		t.Fatal("dismissing the palette did not hand focus back to the rail search")
	}
	if app.settingsCommandPalette.Query() != "" || app.settingsCommandPalette.editor != nil {
		t.Fatal("closed palette kept its query")
	}
}

func TestSettingsPaletteDoesNotOpenOverAnotherDialog(t *testing.T) {
	app := newSettingsPaletteTestApp(t)
	app.settingsDemo = &settingsDemoState{}

	if app.onSettingsPaletteKey(woxui.KeyEvent{Key: "k", Modifiers: settingsPalettePrimaryModifier(), Down: true}) {
		t.Fatal("primary+K was consumed while another settings dialog is open")
	}
	if app.settingsCommandPalette.Open() {
		t.Fatal("palette opened over another settings dialog")
	}
}

func TestSettingsPaletteSelectionMovesAndClamps(t *testing.T) {
	var palette settingsCommandPaletteState
	palette.show(false)

	palette.move(-1, 3)
	if palette.Selected() != 0 {
		t.Fatalf("selection above the first row = %d, want 0", palette.Selected())
	}
	for range 5 {
		palette.move(1, 3)
	}
	if palette.Selected() != 2 {
		t.Fatalf("selection past the last row = %d, want 2", palette.Selected())
	}
	if got := palette.selectedIndex(2); got != 1 {
		t.Fatalf("selection after results shrank = %d, want last row 1", got)
	}
	if got := palette.selectedIndex(0); got != -1 {
		t.Fatalf("selection without results = %d, want -1", got)
	}
	palette.move(1, 0)
	if palette.Selected() != 0 {
		t.Fatalf("selection without results = %d, want 0", palette.Selected())
	}
	palette.move(1, 3)
	palette.setQuery("theme")
	if palette.Selected() != 0 || palette.Query() != "theme" || palette.editor.Text() != "theme" {
		t.Fatalf("query change = %q/%d, want theme with the first row selected", palette.Query(), palette.Selected())
	}
}

func TestSettingsPaletteUsesSettingsSearchRanking(t *testing.T) {
	app := newSettingsPaletteTestApp(t)
	app.openSettingsPalette()
	_ = app.setSettingsPaletteQuery("  proxy ")

	snapshot := app.settingsSnapshot()
	rail := snapshot
	rail.search.Query = woxui.TextEditingState{Text: "proxy"}
	got := app.settingsPaletteResults(snapshot)
	want := app.settingsSearchResults(rail)
	if len(got) == 0 || !reflect.DeepEqual(got, want) {
		t.Fatalf("palette results = %#v, want rail search results %#v", got, want)
	}
}

func TestSettingsPaletteEnterActivatesSelectedResult(t *testing.T) {
	app := newSettingsPaletteTestApp(t)
	app.settingsSearch.SetFocused(true)
	app.openSettingsPalette()
	_ = app.setSettingsPaletteQuery("LangCode")

	results := app.settingsPaletteResults(app.settingsSnapshot())
	target := -1
	for index, result := range results {
		if result.kind == settingsSearchSetting && result.settingKey == "LangCode" {
			target = index
			break
		}
	}
	if target < 0 {
		t.Fatalf("LangCode missing from palette results: %#v", results)
	}
	for range target {
		if !app.onSettingsPaletteKey(woxui.KeyEvent{Key: woxui.KeyArrowDown, Down: true}) {
			t.Fatal("ArrowDown was not handled by the open palette")
		}
	}
	if app.settingsCommandPalette.Selected() != target {
		t.Fatalf("selection = %d, want %d", app.settingsCommandPalette.Selected(), target)
	}
	if !app.onSettingsPaletteKey(woxui.KeyEvent{Key: woxui.KeyEnter, Down: true}) {
		t.Fatal("Enter was not handled by the open palette")
	}
	if app.settingsCommandPalette.Open() {
		t.Fatal("palette stayed open after activation")
	}
	if app.settingTab != "appearance" || app.settingFlash != "built-in:LangCode" {
		t.Fatalf("activation = tab %q highlight %q, want appearance with the LangCode highlight", app.settingTab, app.settingFlash)
	}
	items := settingItemsForSnapshot(app.settingsSnapshot())
	if app.settingRow < 0 || app.settingRow >= len(items) || items[app.settingRow].key != "LangCode" {
		t.Fatalf("selected row = %d, want the LangCode row", app.settingRow)
	}
	if app.settingsSearch.Focused() {
		t.Fatal("activation re-focused the rail search; rail-search navigation leaves focus cleared")
	}
}
