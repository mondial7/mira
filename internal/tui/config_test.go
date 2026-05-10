package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/mira/internal/listing"
)

func TestSaveLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mira", "config.json")
	want := Settings{Theme: ThemeOcean, Borders: BorderDotted, Bionic: false}
	if err := SaveSettings(path, want); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip: got %+v, want %+v", got, want)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope", "config.json")
	got, err := LoadSettings(path)
	if err != nil {
		t.Errorf("missing file should not error, got %v", err)
	}
	if got != DefaultSettings() {
		t.Errorf("missing file should yield DefaultSettings, got %+v", got)
	}
}

func TestLoadCorruptFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := LoadSettings(path)
	if got != DefaultSettings() {
		t.Errorf("corrupt file should yield DefaultSettings, got %+v", got)
	}
}

func TestLoadIgnoresUnknownLabels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	body := []byte(`{"theme":"sunset","borders":"laser","bionic":false}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := LoadSettings(path)
	if got.Theme != ThemeSlate {
		t.Errorf("unknown theme should fall back to ThemeSlate, got %v", got.Theme)
	}
	if got.Borders != BorderFine {
		t.Errorf("unknown borders should fall back to BorderFine, got %v", got.Borders)
	}
	if got.Bionic {
		t.Error("bionic should still honour the explicit false from the file")
	}
}

func TestLoadOmittedFieldsKeepDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	body := []byte(`{"theme":"forest"}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := LoadSettings(path)
	if got.Theme != ThemeForest {
		t.Errorf("theme = %v, want ThemeForest", got.Theme)
	}
	if got.Borders != BorderFine {
		t.Errorf("borders should default to BorderFine, got %v", got.Borders)
	}
	if !got.Bionic {
		t.Error("bionic should default to true when absent")
	}
}

func TestSaveAtomicReplacesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := SaveSettings(path, Settings{Theme: ThemeOcean, Borders: BorderThick, Bionic: true}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := SaveSettings(path, Settings{Theme: ThemeRose, Borders: BorderDotted, Bionic: false}); err != nil {
		t.Fatalf("second save: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf(".tmp should not linger after a successful save, stat err = %v", err)
	}
	got, _ := LoadSettings(path)
	if got.Theme != ThemeRose {
		t.Errorf("second save should win, got theme %v", got.Theme)
	}
}

// withRedirectedConfigPath points configPathFunc at a tempdir for the
// duration of a test so New()/closeSettings() can be exercised without
// touching the user's real config.
func withRedirectedConfigPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	prev := configPathFunc
	configPathFunc = func() (string, error) { return path, nil }
	t.Cleanup(func() { configPathFunc = prev })
	return path
}

func TestNewLoadsPersistedSettings(t *testing.T) {
	path := withRedirectedConfigPath(t)
	if err := SaveSettings(path, Settings{Theme: ThemeOcean, Borders: BorderDotted, Bionic: false}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if m.settings.Theme != ThemeOcean {
		t.Errorf("theme = %v, want ThemeOcean (loaded from disk)", m.settings.Theme)
	}
	if m.settings.Borders != BorderDotted {
		t.Errorf("borders = %v, want BorderDotted", m.settings.Borders)
	}
	if m.settings.Bionic {
		t.Error("bionic should be false (loaded from disk)")
	}
}

func TestCloseSettingsPersistsChanges(t *testing.T) {
	path := withRedirectedConfigPath(t)
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Open overlay, cycle theme once, close.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next.(Model).settingsMode {
		t.Fatal("esc should have closed the overlay")
	}

	got, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings after close: %v", err)
	}
	if got.Theme != ThemeForest {
		t.Errorf("persisted theme = %v, want ThemeForest", got.Theme)
	}
}

func TestCloseSettingsSkipsWriteWhenUnchanged(t *testing.T) {
	path := withRedirectedConfigPath(t)
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Open + close without changing anything.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.settingsMode {
		t.Fatal("esc should have closed the overlay")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("no-op overlay session should not create a config file, stat err = %v", err)
	}
}
