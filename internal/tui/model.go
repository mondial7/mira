package tui

import (
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/mira/internal/listing"
)

// headerLines and footerLines define how many lines the top header block
// and the bottom critter+help block consume. chromeLines is their sum
// (plus blank separators) — the lines NOT available for the grid.
const (
	// Header = 1 path bar + 1 stats summary + 1 blank.
	headerLines = 3
	// Footer = 1 blank gap + CritterHeight + 1 blank + 1 help line.
	footerLines = 1 + CritterHeight + 1 + 1
	// chromeLines = header + footer (their constants already include the
	// blanks that border them). Used to compute how many grid rows fit.
	chromeLines = headerLines + footerLines
)

// tickInterval drives the ambient blink/wag loop. Critters always face
// the column the cursor is in, so we don't need a separate "decay" timer
// for snapping-back-to-centre behaviour.
const tickInterval = 600 * time.Millisecond

// tickMsg is the bubbletea message dispatched on each animation tick.
type tickMsg time.Time

// Model is the bubbletea state for the file browser. It owns the current
// directory, the cached listing, the cursor index, the rendered terminal
// dimensions, and the animation state for the bottom-right critters.
type Model struct {
	cwd     string
	entries []listing.Entry
	cursor  int

	width  int
	height int

	opts listing.Options
	err  error

	// animFrame is a monotonic tick counter; modular cycling done in
	// pickFrames. The critters' look-direction is derived from the cursor
	// position itself, so it's not stored here.
	animFrame int

	// cellW is the per-listing card width, recomputed after every refresh
	// so it expands to fit the longest entry name without truncation.
	cellW int

	// Aggregated stats over the current listing, displayed in the header.
	totalFiles int
	totalDirs  int
	totalSize  int64
	sizeExact  bool

	// QuitWithCD is set when the user pressed the "end here" key (e).
	// The runner inspects this after tea.Quit so a wrapper shell
	// function can capture the path off stdout. (Was bound to capital Q
	// pre-v0.2; the rename came with the cd-handoff rewrite.)
	QuitWithCD bool

	// scrollOffset is the index of the first grid row currently visible
	// in the viewport. It's adjusted automatically as the cursor moves
	// (see ensureCursorVisible) and on window resizes.
	scrollOffset int

	// Fuzzy search state. While searchMode is true, m.entries is the
	// filtered subset and m.fullEntries holds the original listing.
	// Exiting search restores m.entries from m.fullEntries.
	searchMode  bool
	searchQuery string
	fullEntries []listing.Entry

	// User-tunable presentation knobs (theme, borders, bionic toggle).
	// settings is the live state; styles is the lipgloss bundle derived
	// from settings.Theme — re-built whenever the theme changes.
	settings Settings
	styles   themeStyles

	// settingsMode is true while the settings overlay is open. While in
	// this mode, regular navigation keys are intercepted by the settings
	// handler instead of the file browser. settingsCursor tracks which
	// settings row is currently focused.
	settingsMode   bool
	settingsCursor int
}

// New constructs a Model rooted at start. It returns an error only when
// the initial directory cannot be read; subsequent navigation errors are
// surfaced via the model's status line instead.
func New(start string, opts listing.Options) (Model, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return Model{}, err
	}
	m := Model{cwd: abs, opts: opts, settings: DefaultSettings()}
	m.applyTheme()
	if err := m.refresh(); err != nil {
		return Model{}, err
	}
	return m, nil
}

// applyTheme re-builds m.styles from the current settings.Theme. Called
// after any change to settings.Theme (and once at construction).
func (m *Model) applyTheme() {
	m.styles = paletteFor(m.settings.Theme).styles()
}

func (m *Model) refresh() error {
	entries, err := listing.List(m.cwd, m.opts)
	if err != nil {
		return err
	}
	m.entries = entries
	if m.cursor >= m.totalItems() {
		m.cursor = 0
	}
	m.err = nil
	m.recomputeAggregates()
	m.cellW = m.computeCellWidth()
	// New listing → start at the top of the grid.
	m.scrollOffset = 0
	m.ensureCursorVisible()
	return nil
}

// recomputeAggregates folds the listing into the header stats counters.
// sizeExact stays true only if every directory size in the listing was
// fully resolved within the walk budget.
func (m *Model) recomputeAggregates() {
	m.totalDirs = 0
	m.totalFiles = 0
	m.totalSize = 0
	m.sizeExact = true
	for _, e := range m.entries {
		if e.IsDir {
			m.totalDirs++
		} else {
			m.totalFiles++
		}
		m.totalSize += e.Size
		if !e.SizeExact {
			m.sizeExact = false
		}
	}
}

// CWD returns the directory the user was looking at when the program
// exited. Exposed so the binary's --cd mode can print it to stdout for
// shell-wrapper consumption.
func (m Model) CWD() string { return m.cwd }

// totalItems counts entries plus the synthetic ".." at index 0.
func (m Model) totalItems() int { return len(m.entries) + 1 }

func (m Model) isParent(i int) bool { return i == 0 }

func (m Model) entryAt(i int) listing.Entry { return m.entries[i-1] }

// cellWidth returns the active card width, falling back to the minimum
// before refresh has populated cellW.
func (m Model) cellWidth() int {
	if m.cellW < MinCellWidth {
		return MinCellWidth
	}
	return m.cellW
}

// columnStride is the horizontal advance from one card's left edge to
// the next: card width plus the inter-card gap.
func (m Model) columnStride() int { return m.cellWidth() + colGap }

// cols returns the number of cells per row for the current width, never
// less than 1 so the layout stays valid in narrow terminals.
func (m Model) cols() int {
	cw := m.cellWidth()
	if m.width < cw {
		return 1
	}
	stride := m.columnStride()
	// Add colGap before dividing so we count the trailing card whose gap
	// would otherwise overflow the available width.
	return (m.width + colGap) / stride
}

// cellAt maps a terminal click to a cell index. y is in absolute terminal
// coordinates; this function applies the header offset, the active
// scrollOffset, and the inter-card / inter-row gaps.
func (m Model) cellAt(x, y int) int {
	gridY := y - headerLines
	if gridY < 0 {
		return -1
	}
	cols := m.cols()
	stride := m.columnStride()
	col := x / stride
	if x%stride >= m.cellWidth() {
		return -1
	}
	rowStride := CellHeight + rowGap
	visualRow := gridY / rowStride
	if gridY%rowStride >= CellHeight {
		return -1
	}
	if visualRow < 0 || visualRow >= m.visibleGridRows() {
		return -1
	}
	if col < 0 || col >= cols {
		return -1
	}
	row := visualRow + m.scrollOffset
	i := row*cols + col
	if i < 0 || i >= m.totalItems() {
		return -1
	}
	return i
}

// visibleGridRows returns how many full rows of cards fit in the
// terminal beneath the header and above the footer/critters. It clamps
// to at least 1 so we always render something even on absurdly small
// terminals; before the first WindowSizeMsg arrives it returns a large
// number so the snapshot tests (which set width but not height) still
// see all rows.
func (m Model) visibleGridRows() int {
	if m.height <= 0 {
		return 1 << 16
	}
	available := m.height - chromeLines
	if available < CellHeight {
		return 1
	}
	return (available + rowGap) / (CellHeight + rowGap)
}

// totalGridRows is how many rows the listing would occupy if every row
// were rendered.
func (m Model) totalGridRows() int {
	cols := m.cols()
	if cols < 1 {
		return 0
	}
	n := m.totalItems()
	return (n + cols - 1) / cols
}

// cursorRow is the grid row that currently holds the cursor.
func (m Model) cursorRow() int {
	cols := m.cols()
	if cols < 1 {
		return 0
	}
	return m.cursor / cols
}

// ensureCursorVisible nudges scrollOffset so cursorRow sits inside the
// current viewport. Called after every cursor change, refresh, and
// resize.
func (m *Model) ensureCursorVisible() {
	visible := m.visibleGridRows()
	cur := m.cursorRow()
	if cur < m.scrollOffset {
		m.scrollOffset = cur
	} else if cur >= m.scrollOffset+visible {
		m.scrollOffset = cur - visible + 1
	}
	total := m.totalGridRows()
	if m.scrollOffset+visible > total {
		m.scrollOffset = total - visible
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// computeCellWidth measures the longest name + stats string across the
// current listing (including the synthetic ".." entry) and picks a card
// width that fits all of them — clamped to MinCellWidth on the low end
// and the terminal width on the high end so the grid never overflows.
func (m Model) computeCellWidth() int {
	maxInner := MinCellWidth - 2
	for i := 0; i < m.totalItems(); i++ {
		nameLine, statsLine := cardTextLines(m, i)
		w := nameLine
		if statsLine > w {
			w = statsLine
		}
		if w > maxInner {
			maxInner = w
		}
	}
	cellW := maxInner + 3 // 2 borders + 1 char of trailing breathing room
	if m.width > 0 && cellW > m.width {
		cellW = m.width
	}
	if cellW < MinCellWidth {
		cellW = MinCellWidth
	}
	return cellW
}

// cursorLookDir returns -1, 0, or 1 based on which third of the row the
// cursor is currently in. The critters use this to track the user's
// selection without needing any extra state machine.
func (m Model) cursorLookDir() int {
	cols := m.cols()
	if cols <= 1 || m.totalItems() == 0 {
		return 0
	}
	col := m.cursor % cols
	third := cols / 3
	if third < 1 {
		third = 1
	}
	switch {
	case col < third:
		return -1
	case col >= cols-third:
		return 1
	default:
		return 0
	}
}

// Init starts the animation tick loop.
func (Model) Init() tea.Cmd { return tickCmd() }

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update is the bubbletea event handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureCursorVisible()
		return m, nil

	case tickMsg:
		m.animFrame++
		return m, tickCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsMode {
		return m.handleSettingsKey(msg)
	}
	if m.searchMode {
		return m.handleSearchKey(msg)
	}
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "e":
		// End-here: quit and ask the wrapper shell function to cd into
		// the current dir. Replaces the v0.1 capital-Q binding (which
		// suffered from a stdout-handoff bug).
		m.QuitWithCD = true
		return m, tea.Quit
	case "left", "a":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "d":
		if m.cursor < m.totalItems()-1 {
			m.cursor++
		}
	case "up", "w":
		if m.cursor-m.cols() >= 0 {
			m.cursor -= m.cols()
		}
	case "down", "s":
		if m.cursor+m.cols() < m.totalItems() {
			m.cursor += m.cols()
		}
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = m.totalItems() - 1
	case "enter", " ":
		m.activate(m.cursor)
	case "backspace", "esc":
		m.goUp()
	case "h":
		m.toggleHidden()
	case "f":
		m.startSearch()
	case ".":
		m.openSettings()
	}
	m.ensureCursorVisible()
	return m, nil
}

// handleSettingsKey is the modal key handler used while settingsMode is
// on. ↑/↓ moves between rows, ←/→ cycles the focused row's value (and
// enter/space toggles bionic), and "." or esc closes the overlay.
func (m Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", ".":
		m.closeSettings()
	case "up", "w", "k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "down", "s", "j":
		if m.settingsCursor < len(settingsFields)-1 {
			m.settingsCursor++
		}
	case "left", "a", "h":
		m.adjustSetting(-1)
	case "right", "d", "l":
		m.adjustSetting(1)
	case "enter", " ":
		m.adjustSetting(1)
	}
	return m, nil
}

// openSettings turns on the settings overlay, parking the focus on the
// first row so the user always lands somewhere predictable.
func (m *Model) openSettings() {
	m.settingsMode = true
	m.settingsCursor = 0
}

// closeSettings dismisses the overlay, returning to the file browser.
func (m *Model) closeSettings() {
	m.settingsMode = false
}

// adjustSetting cycles the value of the currently focused settings row
// by delta (+1 or -1). Bionic is a toggle so any non-zero delta flips
// it; the other fields wrap through their preset list.
func (m *Model) adjustSetting(delta int) {
	if m.settingsCursor < 0 || m.settingsCursor >= len(settingsFields) {
		return
	}
	switch settingsFields[m.settingsCursor] {
	case settingTheme:
		m.settings.cycleTheme(delta)
		m.applyTheme()
	case settingBorders:
		m.settings.cycleBorders(delta)
	case settingBionic:
		m.settings.toggleBionic()
	}
}

// handleSearchKey is the modal key handler used while searchMode is on.
// Letter keys feed the query (so a/d/w/s/h/q/e lose their nav meaning),
// arrow keys still navigate matches, esc cancels, ctrl+c always quits,
// and enter opens the highlighted match.
func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.endSearch()
	case "enter":
		m.activate(m.cursor)
	case "backspace":
		if r := []rune(m.searchQuery); len(r) > 0 {
			m.searchQuery = string(r[:len(r)-1])
			m.updateFilter()
		}
	case "left":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right":
		if m.cursor < m.totalItems()-1 {
			m.cursor++
		}
	case "up":
		if m.cursor-m.cols() >= 0 {
			m.cursor -= m.cols()
		}
	case "down":
		if m.cursor+m.cols() < m.totalItems() {
			m.cursor += m.cols()
		}
	default:
		if isPrintable(msg) {
			m.searchQuery += string(msg.Runes)
			m.updateFilter()
		}
	}
	m.ensureCursorVisible()
	return m, nil
}

// isPrintable reports whether a key message represents typed character
// input (any rune ≥ space, no control codes, no non-rune events).
func isPrintable(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 {
		return false
	}
	for _, r := range msg.Runes {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

// startSearch enters fuzzy-search mode, snapshotting the current
// listing so it can be restored on cancel. The cursor resets to 0 so
// the first match (after any keystroke) is the initial selection.
func (m *Model) startSearch() {
	if m.searchMode {
		return
	}
	m.searchMode = true
	m.searchQuery = ""
	m.fullEntries = m.entries
	m.cursor = 0
	m.scrollOffset = 0
}

// endSearch exits search mode and restores the un-filtered listing.
// Safe to call when not in search mode (no-op).
func (m *Model) endSearch() {
	if !m.searchMode {
		return
	}
	m.searchMode = false
	m.searchQuery = ""
	if m.fullEntries != nil {
		m.entries = m.fullEntries
		m.fullEntries = nil
	}
	m.cursor = 0
	m.scrollOffset = 0
	m.recomputeAggregates()
	m.cellW = m.computeCellWidth()
}

// updateFilter re-runs the fuzzy match against the current query and
// replaces m.entries with the matching subset. Called after every
// keystroke that changes m.searchQuery.
func (m *Model) updateFilter() {
	if !m.searchMode {
		return
	}
	filtered := make([]listing.Entry, 0, len(m.fullEntries))
	for _, e := range m.fullEntries {
		if fuzzyMatch(m.searchQuery, e.Name) {
			filtered = append(filtered, e)
		}
	}
	m.entries = filtered
	if m.cursor >= m.totalItems() {
		m.cursor = 0
	}
	m.scrollOffset = 0
	m.recomputeAggregates()
	m.cellW = m.computeCellWidth()
}

// fuzzyMatch reports whether every rune of query appears in target in
// order, case-insensitively. An empty query matches everything.
func fuzzyMatch(query, target string) bool {
	qr := []rune(strings.ToLower(query))
	if len(qr) == 0 {
		return true
	}
	qi := 0
	for _, c := range strings.ToLower(target) {
		if c == qr[qi] {
			qi++
			if qi == len(qr) {
				return true
			}
		}
	}
	return false
}

// toggleHidden flips the ShowHidden option and re-lists the directory.
// Errors during refresh are surfaced through the model's status line so
// the user sees them without losing the current navigation context.
func (m *Model) toggleHidden() {
	m.opts.ShowHidden = !m.opts.ShowHidden
	if err := m.refresh(); err != nil {
		m.err = err
	}
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	i := m.cellAt(msg.X, msg.Y)
	if i < 0 {
		return m, nil
	}
	m.cursor = i
	m.activate(i)
	m.ensureCursorVisible()
	return m, nil
}

// activate opens the item at i: the synthetic ".." goes up; directories
// descend; regular files are no-ops (kept for future preview support).
// Search state is cleared on descent — the new directory has its own
// listing and the user almost certainly wants to see all of it.
func (m *Model) activate(i int) {
	if m.isParent(i) {
		m.goUp()
		return
	}
	e := m.entryAt(i)
	if !e.IsDir {
		return
	}

	// Drop any active search before refresh; refresh() repopulates entries.
	m.searchMode = false
	m.searchQuery = ""
	m.fullEntries = nil

	prev := m.cwd
	m.cwd = e.Path
	m.cursor = 0
	if err := m.refresh(); err != nil {
		m.cwd = prev
		m.err = err
	}
}

func (m *Model) goUp() {
	parent := filepath.Dir(m.cwd)
	if parent == m.cwd {
		return
	}
	m.cwd = parent
	m.cursor = 0
	if err := m.refresh(); err != nil {
		m.err = err
	}
}
