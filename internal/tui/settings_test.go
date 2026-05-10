package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/mira/internal/listing"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s.Theme != ThemeSlate {
		t.Errorf("default theme = %v, want ThemeSlate", s.Theme)
	}
	if s.Borders != BorderFine {
		t.Errorf("default borders = %v, want BorderFine", s.Borders)
	}
	if !s.Bionic {
		t.Error("default Bionic should be true")
	}
}

func TestColorTheme_Labels(t *testing.T) {
	cases := map[ColorTheme]string{
		ThemeSlate:  "slate",
		ThemeForest: "forest",
		ThemeOcean:  "ocean",
		ThemeRose:   "rose",
	}
	for tt, want := range cases {
		if got := tt.Label(); got != want {
			t.Errorf("Label(%v) = %q, want %q", tt, got, want)
		}
	}
}

func TestBorderPreset_Labels(t *testing.T) {
	cases := map[BorderPreset]string{
		BorderFine:   "fine",
		BorderThick:  "thick",
		BorderDotted: "dotted",
	}
	for b, want := range cases {
		if got := b.Label(); got != want {
			t.Errorf("Label(%v) = %q, want %q", b, got, want)
		}
	}
}

func TestSettings_CycleTheme_Wraps(t *testing.T) {
	s := DefaultSettings() // ThemeSlate
	s.cycleTheme(1)
	if s.Theme != ThemeForest {
		t.Errorf("after +1 theme = %v, want ThemeForest", s.Theme)
	}
	// Wrap forward all the way back to slate.
	for i := 0; i < len(allThemes)-1; i++ {
		s.cycleTheme(1)
	}
	if s.Theme != ThemeSlate {
		t.Errorf("after full forward cycle theme = %v, want ThemeSlate", s.Theme)
	}
	// Wrap backward through zero.
	s.cycleTheme(-1)
	if s.Theme != ThemeRose {
		t.Errorf("after -1 from slate theme = %v, want ThemeRose", s.Theme)
	}
}

func TestSettings_CycleBorders_Wraps(t *testing.T) {
	s := DefaultSettings()
	s.cycleBorders(1)
	if s.Borders != BorderThick {
		t.Errorf("after +1 borders = %v, want BorderThick", s.Borders)
	}
	s.cycleBorders(1)
	if s.Borders != BorderDotted {
		t.Errorf("after +2 borders = %v, want BorderDotted", s.Borders)
	}
	s.cycleBorders(1)
	if s.Borders != BorderFine {
		t.Errorf("after wrap borders = %v, want BorderFine", s.Borders)
	}
}

func TestSettings_ToggleBionic(t *testing.T) {
	s := DefaultSettings()
	if !s.Bionic {
		t.Fatal("default Bionic should be true")
	}
	s.toggleBionic()
	if s.Bionic {
		t.Error("after toggle Bionic should be false")
	}
	s.toggleBionic()
	if !s.Bionic {
		t.Error("after second toggle Bionic should be true again")
	}
}

func TestPickGlyphs(t *testing.T) {
	cases := []struct {
		preset   BorderPreset
		kind     cardKind
		selected bool
		want     glyphSet
		desc     string
	}{
		{BorderFine, kindDir, false, glyphsRound, "fine + dir → rounded"},
		{BorderFine, kindFile, false, glyphsDouble, "fine + file → double"},
		{BorderFine, kindDir, true, glyphsHeavy, "fine + selected → heavy"},
		{BorderFine, kindFile, true, glyphsHeavy, "fine + selected file → heavy"},
		{BorderThick, kindDir, false, glyphsHeavy, "thick → heavy always"},
		{BorderThick, kindFile, false, glyphsHeavy, "thick file → heavy"},
		{BorderThick, kindDir, true, glyphsHeavy, "thick selected → heavy"},
		{BorderDotted, kindDir, false, glyphsDotted, "dotted dir"},
		{BorderDotted, kindFile, false, glyphsDotted, "dotted file"},
		{BorderDotted, kindDir, true, glyphsDotted, "dotted selected"},
	}
	for _, tc := range cases {
		got := pickGlyphs(tc.preset, tc.kind, tc.selected)
		if got != tc.want {
			t.Errorf("%s: got %+v, want %+v", tc.desc, got, tc.want)
		}
	}
}

func TestModel_DotKeyOpensSettings(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if m.settingsMode {
		t.Fatal("settingsMode should default off")
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	if !m.settingsMode {
		t.Fatal(`"." should open settings`)
	}
	if m.settingsCursor != 0 {
		t.Errorf("settingsCursor should reset to 0, got %d", m.settingsCursor)
	}
}

func TestModel_SettingsEscClosesOverlay(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.settingsMode {
		t.Error("esc should close settings")
	}
}

func TestModel_SettingsDotClosesOverlay(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	if m.settingsMode {
		t.Error(`pressing "." again should toggle settings off`)
	}
}

func TestModel_SettingsRightCyclesTheme(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Open settings; cursor lands on theme row.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)

	if m.settings.Theme != ThemeSlate {
		t.Fatalf("starting theme = %v, want ThemeSlate", m.settings.Theme)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(Model)
	if m.settings.Theme != ThemeForest {
		t.Errorf("after right theme = %v, want ThemeForest", m.settings.Theme)
	}
}

func TestModel_SettingsBordersAndBionic(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	// Move cursor down to borders row.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(Model)
	if m.settings.Borders != BorderThick {
		t.Errorf("borders = %v, want BorderThick", m.settings.Borders)
	}

	// Move down to bionic row, toggle off.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	if !m.settings.Bionic {
		t.Fatal("bionic should still be true before toggle")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.settings.Bionic {
		t.Error("enter on bionic row should toggle it off")
	}
}

func TestModel_SettingsViewMentionsCurrentValues(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = 80
	m.height = 30
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	out := m.View()
	for _, want := range []string{"Settings", "Theme", "Borders", "Bionic", "slate", "fine", "on"} {
		if !strings.Contains(out, want) {
			t.Errorf("settings view missing %q\n%s", want, out)
		}
	}
}

func TestModel_BionicOffSkipsBoldRuns(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.settings.Bionic = false
	m.width = 200
	m.height = 30

	out := m.View()
	// scaffoldDir writes "alpha" — when bionic is off, the leading "al"
	// should NOT appear inside a bold-prefix ANSI marker. Spot-check the
	// most obvious one.
	if strings.Contains(out, "\x1b[1;38;5;250mal\x1b[0m") {
		t.Errorf("bionic disabled should not bold leading-half segments\n%s", out)
	}
}

func TestModel_ApplyThemeSwapsStyles(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	before := m.styles.path.GetBackground()
	m.settings.Theme = ThemeOcean
	m.applyTheme()
	after := m.styles.path.GetBackground()
	if before == after {
		t.Errorf("theme swap should change path background colour, got %v == %v", before, after)
	}
}
