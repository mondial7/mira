package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/mondial7/mira/internal/listing"
)

func TestClampDisplay(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"", 5, ""},
		{"hi", 5, "hi"},
		{"hello world", 5, "hell…"},
		{"abc", 1, "…"},
		{"abc", 0, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "ab…"},
	}
	for _, tc := range cases {
		if got := clampDisplay(tc.in, tc.max); got != tc.want {
			t.Errorf("clampDisplay(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestPadToWidth_PadsToExactWidth(t *testing.T) {
	got := padToWidth("ab", 6, defaultStyles.name)
	if w := lipgloss.Width(got); w != 6 {
		t.Errorf("display width = %d, want 6 (%q)", w, got)
	}
	if !strings.Contains(got, "ab") {
		t.Errorf("content lost, got %q", got)
	}
}

func TestStyleName_BionicBoldsLeadingHalfOfWords(t *testing.T) {
	// Force ANSI emission so we can inspect the result.
	prev := lipgloss.DefaultRenderer().ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	lipgloss.SetColorProfile(termenv.ANSI256)

	base := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	cases := []struct {
		name      string
		wantBold  []string // segments that should appear bolded
		wantPlain []string // segments that should appear plain
	}{
		{name: "documents", wantBold: []string{"docu"}, wantPlain: []string{"ments"}},
		{name: "src", wantBold: []string{"sr"}, wantPlain: []string{"c"}},
		{name: "main.go", wantBold: []string{"ma", "g"}, wantPlain: []string{"in", "o"}},
		{name: "node_modules", wantBold: []string{"no", "modu"}, wantPlain: []string{"de", "_", "les"}},
	}
	for _, tc := range cases {
		got := styleName(tc.name, base, true)
		for _, seg := range tc.wantBold {
			marker := "\x1b[1;" // bold attribute, then color
			if !strings.Contains(got, marker+"38;5;250m"+seg) {
				t.Errorf("styleName(%q): expected %q to be bolded, got %q",
					tc.name, seg, got)
			}
		}
	}

	// Bionic disabled passes through to a single styled run.
	plain := styleName("documents", base, false)
	if strings.Contains(plain, "\x1b[1;") {
		t.Errorf("bionic=false should not emit bold, got %q", plain)
	}
}

func TestStyleName_EmptyAndSingleChar(t *testing.T) {
	base := lipgloss.NewStyle()
	if got := styleName("", base, true); got != "" {
		t.Errorf("empty name should render empty, got %q", got)
	}
	// Single char with bionic should still render (the whole 1-char word
	// becomes bold, which is fine).
	if got := styleName("x", base, true); got == "" {
		t.Errorf("single-char name should not be empty")
	}
}

func TestChildCountLabel(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{-1, "—"},
		{0, "empty"},
		{1, "1 item"},
		{42, "42 items"},
	}
	for _, tc := range cases {
		if got := childCountLabel(tc.n); got != tc.want {
			t.Errorf("childCountLabel(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// scaffoldForView builds a tiny fixture for end-to-end View rendering.
func scaffoldForView(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return root
}

// TestView_RendersAllSections is a smoke test: it doesn't assert exact ANSI
// output (that would be brittle), only that View produces non-empty output
// containing the path, every entry name, and the help text.
func TestView_RendersAllSections(t *testing.T) {
	root := scaffoldForView(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Use a generous width so the path doesn't get truncated by renderHeader.
	m.width = 200
	m.height = 30

	out := m.View()
	if out == "" {
		t.Fatal("View() returned empty string")
	}
	// scaffoldForView creates 1 dir + 1 file → "1 folder" + "1 file" in
	// the header summary; "quit" is the always-present footer hint.
	for _, want := range []string{filepath.Base(root), "src", "go.mod", "1 folder", "1 file", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q\n%s", want, out)
		}
	}
}

func TestView_EmptyDirectoryShowsHint(t *testing.T) {
	root := t.TempDir()
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 3
	out := m.View()
	// The grid is empty (only ".." parent), so the help text should still appear.
	if !strings.Contains(out, "..") {
		t.Errorf("expected '..' parent in output:\n%s", out)
	}
}

func TestFlatList_FormatsByEntryType(t *testing.T) {
	entries := []listing.Entry{
		{Name: "src", IsDir: true},
		{Name: "go.mod"},
		{Name: "link", IsSymlink: true, Target: "go.mod"},
	}
	got := FlatList(entries)
	want := "src/\ngo.mod\nlink -> go.mod\n"
	if got != want {
		t.Errorf("FlatList:\n got: %q\nwant: %q", got, want)
	}
}
