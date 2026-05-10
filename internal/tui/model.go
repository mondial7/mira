package tui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marcomondini/banana-four/internal/listing"
)

// headerLines and footerLines are subtracted from terminal height when
// computing how many grid rows fit. They also tell the click handler how
// many lines to skip when mapping y → row.
const (
	headerLines = 2 // path bar + blank line
	footerLines = 2 // blank line + help bar
)

// Model is the bubbletea state for the file browser. It owns the current
// directory, the cached listing, the cursor index and the rendered terminal
// dimensions. The cursor index 0 is reserved for the synthetic ".." entry.
type Model struct {
	cwd     string
	entries []listing.Entry
	cursor  int

	width  int
	height int

	opts listing.Options
	err  error
}

// New constructs a Model rooted at start. It returns an error only when the
// initial directory cannot be read; subsequent navigation errors are surfaced
// via the model's status line instead.
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
	return nil
}

// totalItems counts entries plus the synthetic ".." at index 0.
func (m Model) totalItems() int { return len(m.entries) + 1 }

func (m Model) isParent(i int) bool { return i == 0 }

func (m Model) entryAt(i int) listing.Entry { return m.entries[i-1] }

// cols returns the number of cells per row for the current width, never
// less than 1 so the layout stays valid in narrow terminals.
func (m Model) cols() int {
	if m.width < CellWidth {
		return 1
	}
	return m.width / CellWidth
}

// cellAt maps a terminal click to a cell index. y is in absolute terminal
// coordinates; this function applies the header offset itself.
func (m Model) cellAt(x, y int) int {
	gridY := y - headerLines
	if gridY < 0 {
		return -1
	}
	cols := m.cols()
	col := x / CellWidth
	row := gridY / CellHeight
	if col < 0 || col >= cols {
		return -1
	}
	i := row*cols + col
	if i < 0 || i >= m.totalItems() {
		return -1
	}
	return i
}

// Init satisfies tea.Model; nothing to bootstrap.
func (Model) Init() tea.Cmd { return nil }

// Update is the bubbletea event handler. It returns the new model and
// optionally a command (e.g. tea.Quit).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor < m.totalItems()-1 {
			m.cursor++
		}
	case "up", "k":
		if m.cursor-m.cols() >= 0 {
			m.cursor -= m.cols()
		}
	case "down", "j":
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
	}
	return m, nil
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
