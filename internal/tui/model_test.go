package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/banana-four/internal/listing"
)

// scaffoldDir creates a tmp dir with two subdirs and two files so we can
// drive Model logic deterministically.
func scaffoldDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"alpha", "beta"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	for _, f := range []string{"one.txt", "two.txt"} {
		if err := os.WriteFile(filepath.Join(root, f), nil, 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}
	return root
}

func TestNew_FailsOnUnreadableDir(t *testing.T) {
	_, err := New("/this/path/should/not/exist/banana", listing.Options{})
	if err == nil {
		t.Fatal("expected error from New on missing directory")
	}
}

func TestModel_TotalItemsIncludesParent(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// 2 dirs + 2 files + 1 parent = 5
	if got := m.totalItems(); got != 5 {
		t.Errorf("totalItems = %d, want 5", got)
	}
}

func TestModel_ColsClampsToOne(t *testing.T) {
	m := Model{width: 4} // less than CellWidth
	if got := m.cols(); got != 1 {
		t.Errorf("cols = %d, want 1 for narrow width", got)
	}
}

func TestModel_ColsScalesWithWidth(t *testing.T) {
	// With cellW unset, cellWidth() returns MinCellWidth. Stride is
	// MinCellWidth + colGap; we want 3 cards to fit exactly.
	stride := MinCellWidth + colGap
	m := Model{width: stride*3 - colGap}
	if got := m.cols(); got != 3 {
		t.Errorf("cols = %d, want 3", got)
	}
}

func TestModel_CellAtMapsClicksToIndex(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	stride := m.columnStride()
	m.width = stride*3 - colGap // 3 columns

	// First cell of first row → ".." (index 0).
	if got := m.cellAt(1, headerLines); got != 0 {
		t.Errorf("first cell click = %d, want 0", got)
	}
	// Second cell of first row starts at columnStride.
	if got := m.cellAt(stride+1, headerLines); got != 1 {
		t.Errorf("second cell click = %d, want 1", got)
	}
	// A click in the gutter between cards is rejected.
	if got := m.cellAt(m.cellWidth(), headerLines); got != -1 {
		t.Errorf("gutter click = %d, want -1", got)
	}
	// Click in the header row (above grid).
	if got := m.cellAt(0, 0); got != -1 {
		t.Errorf("header click = %d, want -1", got)
	}
	// Click way past last item.
	if got := m.cellAt(0, headerLines+(CellHeight+rowGap)*99); got != -1 {
		t.Errorf("out-of-range click = %d, want -1", got)
	}
}

func TestModel_KeyboardNavigationStaysInBounds(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5

	// Right past the end should clamp.
	for i := 0; i < 100; i++ {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = next.(Model)
	}
	if m.cursor != m.totalItems()-1 {
		t.Errorf("cursor = %d, want last index %d", m.cursor, m.totalItems()-1)
	}

	// Left past the start should clamp.
	for i := 0; i < 100; i++ {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
		m = next.(Model)
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestModel_EnterDescendsIntoDirectory(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5

	// First non-parent entry should be "alpha" (dirs sort before files).
	m.cursor = 1
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)

	if filepath.Base(m.cwd) != "alpha" {
		t.Errorf("cwd = %q, want to end in alpha", m.cwd)
	}
	if m.cursor != 0 {
		t.Errorf("cursor reset expected after navigation, got %d", m.cursor)
	}
}

func TestModel_BackspaceGoesUp(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(filepath.Join(root, "alpha"), listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(Model)

	if m.cwd != root {
		t.Errorf("cwd after backspace = %q, want %q", m.cwd, root)
	}
}

func TestModel_CursorLookDirTracksPosition(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Force a 3-column layout so the thirds carve cleanly. We pull the
	// stride from the model so dynamic cell sizing (which depends on the
	// actual entries) doesn't break the test.
	stride := m.columnStride()
	m.width = stride*3 - colGap

	cases := []struct {
		cursor int
		want   int
	}{
		{0, -1}, // leftmost column
		{1, 0},  // middle
		{2, 1},  // rightmost
		{3, -1}, // wraps to next row, leftmost
	}
	for _, tc := range cases {
		m.cursor = tc.cursor
		if got := m.cursorLookDir(); got != tc.want {
			t.Errorf("cursor=%d → lookDir=%d, want %d", tc.cursor, got, tc.want)
		}
	}
}

func TestModel_AnimFrameAdvancesOnTick(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	startFrame := m.animFrame
	next, _ := m.Update(tickMsg{})
	if next.(Model).animFrame != startFrame+1 {
		t.Errorf("animFrame = %d, want %d", next.(Model).animFrame, startFrame+1)
	}
}

// scaffoldManyDirs creates a tmp directory containing 30 numbered
// subdirectories so we can exercise the scrolling code paths.
func scaffoldManyDirs(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for i := 0; i < 30; i++ {
		name := filepath.Join(root, fmt.Sprintf("dir%02d", i))
		if err := os.MkdirAll(name, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	return root
}

func TestModel_VisibleGridRowsScalesWithHeight(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Big terminal: at least 3 rows fit.
	m.height = 60
	if got := m.visibleGridRows(); got < 3 {
		t.Errorf("visibleGridRows(60) = %d, want >= 3", got)
	}
	// Tiny terminal: never returns 0.
	m.height = 5
	if got := m.visibleGridRows(); got < 1 {
		t.Errorf("visibleGridRows(5) = %d, want >= 1", got)
	}
}

func TestModel_ScrollFollowsCursorDown(t *testing.T) {
	root := scaffoldManyDirs(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Force a tight viewport: 3 columns × 2 visible rows = 6 visible items.
	m.width = m.columnStride()*3 - colGap
	m.height = chromeLines + 2*(CellHeight+rowGap) - rowGap // exactly 2 rows
	m.ensureCursorVisible()

	if m.scrollOffset != 0 {
		t.Fatalf("initial scrollOffset = %d, want 0", m.scrollOffset)
	}

	// Move cursor down past the visible window.
	for i := 0; i < 10; i++ {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = next.(Model)
	}
	if m.scrollOffset == 0 {
		t.Errorf("scrollOffset still 0 after moving down 10 rows; cursor=%d", m.cursor)
	}
}

func TestModel_ScrollFollowsCursorUp(t *testing.T) {
	root := scaffoldManyDirs(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = m.columnStride()*3 - colGap
	m.height = chromeLines + 2*(CellHeight+rowGap) - rowGap

	// Start near the bottom.
	m.cursor = m.totalItems() - 1
	m.ensureCursorVisible()
	bottomScroll := m.scrollOffset
	if bottomScroll == 0 {
		t.Fatal("expected non-zero scrollOffset when cursor is at bottom")
	}

	// Now jump back to the top with 'g'.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = next.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset after 'g' = %d, want 0", m.scrollOffset)
	}
}

func TestModel_ScrollIndicatorReflectsState(t *testing.T) {
	root := scaffoldManyDirs(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = m.columnStride()*3 - colGap
	m.height = chromeLines + 2*(CellHeight+rowGap) - rowGap

	// At the top: only ▼ (or empty if everything fits — large enough listing
	// here that scrollable is true).
	if got := m.scrollIndicator(); got == "" || !strings.Contains(got, "▼") {
		t.Errorf("indicator at top = %q, want ▼", got)
	}
	if strings.Contains(m.scrollIndicator(), "▲") {
		t.Errorf("indicator at top should not show ▲, got %q", m.scrollIndicator())
	}

	// Jump to the end.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = next.(Model)
	if got := m.scrollIndicator(); !strings.Contains(got, "▲") {
		t.Errorf("indicator at end = %q, want ▲", got)
	}
	if strings.Contains(m.scrollIndicator(), "▼") {
		t.Errorf("indicator at end should not show ▼, got %q", m.scrollIndicator())
	}
}

func TestModel_QuitKey(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd from 'q' key")
	}
	if next.(Model).QuitWithCD {
		t.Error("plain 'q' must not set QuitWithCD")
	}
}

func TestModel_QuitWithCDKey(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd from 'Q' key")
	}
	if !next.(Model).QuitWithCD {
		t.Error("'Q' must set QuitWithCD")
	}
	if next.(Model).CWD() == "" {
		t.Error("CWD() should be populated after navigation")
	}
}

func TestModel_HToggleHiddenFlipsOption(t *testing.T) {
	root := scaffoldDir(t)
	// Add a dotfile so the toggle has something to flip in/out.
	if err := os.WriteFile(filepath.Join(root, ".secret"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if m.opts.ShowHidden {
		t.Fatal("ShowHidden should default to false")
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = next.(Model)
	if !m.opts.ShowHidden {
		t.Fatal("'h' should flip ShowHidden true")
	}

	// Hidden file should now appear in entries.
	found := false
	for _, e := range m.entries {
		if e.Name == ".secret" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(".secret not in entries after toggle: %+v", m.entries)
	}

	// Toggle back off; .secret should disappear.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = next.(Model)
	for _, e := range m.entries {
		if e.Name == ".secret" {
			t.Errorf(".secret still showing after second toggle")
		}
	}
}

func TestModel_WASDNavigation(t *testing.T) {
	root := scaffoldDir(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = next.(Model)
	if m.cursor != 1 {
		t.Errorf("'d' should move right; cursor = %d, want 1", m.cursor)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = next.(Model)
	if m.cursor != 0 {
		t.Errorf("'a' should move left; cursor = %d, want 0", m.cursor)
	}
}
