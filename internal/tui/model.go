package tui

import (
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/banana-four/internal/listing"
)

// headerLines and footerLines define how many lines the top header block
// and the bottom critter+help block consume. They're used for click →
// cell mapping (header = number of lines BEFORE the grid starts).
const (
	// Header = 1 path bar + 1 stats summary + 1 blank.
	headerLines = 3
	// Footer = 1 blank gap + CritterHeight + 1 blank + 1 help line.
	footerLines = 1 + CritterHeight + 1 + 1
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

	// QuitWithCD is set when the user pressed the "quit and cd here" key
	// (capital Q). The runner inspects this after tea.Quit so a wrapper
	// shell function can capture the path off stdout.
	QuitWithCD bool
}

// New constructs a Model rooted at start. It returns an error only when
// the initial directory cannot be read; subsequent navigation errors are
// surfaced via the model's status line instead.
func New(start string, opts listing.Options) (Model, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return Model{}, err
	}
	m := Model{cwd: abs, opts: opts}
	if err := m.refresh(); err != nil {
		return Model{}, err
	}
	return m, nil
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
// coordinates; this function applies the header offset and accounts for
// blank gaps between rows and columns.
func (m Model) cellAt(x, y int) int {
	gridY := y - headerLines
	if gridY < 0 {
		return -1
	}
	cols := m.cols()
	stride := m.columnStride()
	col := x / stride
	// Reject clicks landing in the inter-card horizontal gap.
	if x%stride >= m.cellWidth() {
		return -1
	}
	rowStride := CellHeight + rowGap
	row := gridY / rowStride
	if gridY%rowStride >= CellHeight {
		return -1
	}
	if col < 0 || col >= cols {
		return -1
	}
	i := row*cols + col
	if i < 0 || i >= m.totalItems() {
		return -1
	}
	return i
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
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "Q":
		// Quit and ask the wrapper shell function to cd into the current dir.
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
	}
	return m, nil
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
	return m, nil
}

// activate opens the item at i: the synthetic ".." goes up; directories
// descend; regular files are no-ops (kept for future preview support).
func (m *Model) activate(i int) {
	if m.isParent(i) {
		m.goUp()
		return
	}
	e := m.entryAt(i)
	if !e.IsDir {
		return
	}
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
