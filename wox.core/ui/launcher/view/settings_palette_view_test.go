package view

// KobeTools fork: ⌘K / Ctrl+K settings command palette.

import (
	"testing"

	woxcomponent "wox/ui/launcher/component"
	woxwidget "wox/ui/widget"
)

func TestSettingsCommandPalettePanelSitsInUpperMiddle(t *testing.T) {
	left, top, width := SettingsCommandPalettePanelFrame(1000, 700)
	if width != SettingsCommandPaletteMaxWidth || left != (1000-width)/2 {
		t.Fatalf("panel frame = left %v width %v, want a centered %vpt panel", left, width, SettingsCommandPaletteMaxWidth)
	}
	if top <= 0 || top >= 700/3 {
		t.Fatalf("panel top = %v, want the upper third of the window", top)
	}
	if _, _, narrow := SettingsCommandPalettePanelFrame(400, 700); narrow >= 400 {
		t.Fatalf("narrow window panel width = %v, want a margin inside 400", narrow)
	}
}

func TestSettingsCommandPaletteScrimDismisses(t *testing.T) {
	dismissed := false
	palette := SettingsCommandPalette(SettingsCommandPaletteProps{
		Width: 1000, Height: 700, Query: "theme", Theme: woxcomponent.ControlTheme{},
		Results:   []SettingsSearchResult{{Title: "Themes"}, {Title: "Theme Editor"}},
		OnDismiss: func() { dismissed = true },
	}).(woxwidget.Stack)

	if len(palette.Children) != 2 || palette.Width != 1000 || palette.Height != 700 {
		t.Fatalf("palette layers = %#v, want a full-window scrim and one panel", palette)
	}
	scrim := palette.Children[0].Child.(woxwidget.Semantics).Child.(woxwidget.Gesture)
	if scrim.Child.(woxwidget.Container).Color.A == 0 {
		t.Fatal("scrim is not dimmed")
	}
	scrim.OnTap()
	if !dismissed {
		t.Fatal("scrim tap did not dismiss the palette")
	}
	left, top, _ := SettingsCommandPalettePanelFrame(1000, 700)
	if palette.Children[1].Left != left || palette.Children[1].Top != top {
		t.Fatalf("panel position = %v,%v, want %v,%v", palette.Children[1].Left, palette.Children[1].Top, left, top)
	}
}
