package tui

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// configFileName lives under os.UserConfigDir()/mira/.
const configFileName = "config.json"

// configPathFunc resolves the on-disk location of the persisted settings
// file. Tests redirect this to a tempdir so the user's real config is
// never touched.
var configPathFunc = DefaultConfigPath

// wireConfig is the on-disk JSON shape. Theme/Borders are strings
// (matching their Label()) so reordering the underlying enum doesn't
// invalidate user configs. Bionic is a *bool so we can distinguish
// "absent" (preserve default) from "explicitly false".
type wireConfig struct {
	Theme   string `json:"theme,omitempty"`
	Borders string `json:"borders,omitempty"`
	Bionic  *bool  `json:"bionic,omitempty"`
}

// DefaultConfigPath returns the canonical path for the persisted
// overlay choices. It resolves to os.UserConfigDir()/mira/config.json on
// every supported platform.
func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mira", configFileName), nil
}

// LoadSettings reads the persisted overlay choices from path. A missing
// file or any parse error returns DefaultSettings so the TUI always
// launches; the returned error is non-nil only for genuine I/O failures
// the caller might want to log.
func LoadSettings(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultSettings(), nil
		}
		return DefaultSettings(), err
	}
	var w wireConfig
	if err := json.Unmarshal(data, &w); err != nil {
		return DefaultSettings(), nil
	}
	s := DefaultSettings()
	if t, ok := parseTheme(w.Theme); ok {
		s.Theme = t
	}
	if b, ok := parseBorders(w.Borders); ok {
		s.Borders = b
	}
	if w.Bionic != nil {
		s.Bionic = *w.Bionic
	}
	return s, nil
}

// SaveSettings writes the overlay choices to path atomically: marshal
// to a sibling .tmp file, then rename. A crash mid-write therefore
// cannot leave a half-written config behind.
func SaveSettings(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	bionic := s.Bionic
	w := wireConfig{
		Theme:   s.Theme.Label(),
		Borders: s.Borders.Label(),
		Bionic:  &bionic,
	}
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func parseTheme(s string) (ColorTheme, bool) {
	switch s {
	case "slate":
		return ThemeSlate, true
	case "forest":
		return ThemeForest, true
	case "ocean":
		return ThemeOcean, true
	case "rose":
		return ThemeRose, true
	}
	return 0, false
}

func parseBorders(s string) (BorderPreset, bool) {
	switch s {
	case "fine":
		return BorderFine, true
	case "thick":
		return BorderThick, true
	case "dotted":
		return BorderDotted, true
	}
	return 0, false
}
