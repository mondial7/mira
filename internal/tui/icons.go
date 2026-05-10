package tui

import "github.com/charmbracelet/lipgloss"

// Card geometry. Cells are 6 lines tall (top border, name, separator,
// stats, blank, bottom border) plus one blank line of vertical padding
// between rows. The math elsewhere uses CellHeight (the card itself) and
// rowGap (the spacer) so click-to-cell mapping stays consistent.
const (
	CellWidth  = 20
	CellHeight = 6
	rowGap     = 1 // blank line drawn between rows of cards
	colGap     = 2 // blank columns drawn between cards in a row
)

// Per-row line indices, named for readability.
const (
	lineTop = iota
	lineName
	lineSep
	lineStats
	lineSpacer
	lineBottom
)

// Selection symbols. Selected variants use a different glyph on purpose:
// the change is visible even in the rare environments where lipgloss color
// gets stripped.
const (
	symFolder         = "▸"
	symFolderSelected = "▶"
	symFile           = "·"
	symFileSelected   = "◆"
	symLink           = "↪"
	symLinkSelected   = "⇒"
	symParent         = "↑"
	symParentSelected = "▲"
)

// Border glyphs.
const (
	borderTL, borderTR, borderBL, borderBR = "╭", "╮", "╰", "╯"
	borderH, borderV                       = "─", "│"

	dashedTL, dashedTR, dashedBL, dashedBR = "╭", "╮", "╰", "╯"
	dashedH, dashedV                       = "┈", "┊"

	heavyTL, heavyTR, heavyBL, heavyBR = "┏", "┓", "┗", "┛"
	heavyH, heavyV                     = "━", "┃"

	separatorH = "─" // light horizontal rule beneath the name line
)

// Monochromatic slate-gray palette + a single warm amber accent. Indices
// are 256-color codes; they degrade gracefully on 16-color terminals.
var (
	colDim      = lipgloss.Color("240") // dim border
	colMid      = lipgloss.Color("245") // stats text
	colName     = lipgloss.Color("250") // entry name
	colNameSel  = lipgloss.Color("231") // selected name (bright white)
	colAccent   = lipgloss.Color("214") // amber — used for selected border + symbol
	colCritters = lipgloss.Color("180") // warm sandy color for critters
	colHelp     = lipgloss.Color("242") // help text
	colPathBg   = lipgloss.Color("236") // path bar background
	colPathFg   = lipgloss.Color("231") // path bar text
	colErr      = lipgloss.Color("203") // errors
)

var (
	pathStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colPathFg).
			Background(colPathBg).
			Padding(0, 1)

	borderStyle        = lipgloss.NewStyle().Foreground(colDim)
	selectedStyle      = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	nameStyle          = lipgloss.NewStyle().Foreground(colName)
	nameSelectedStyle  = lipgloss.NewStyle().Foreground(colNameSel).Bold(true)
	statsStyle         = lipgloss.NewStyle().Foreground(colMid)
	statsSelectedStyle = lipgloss.NewStyle().Foreground(colMid).Bold(true)
	critterStyle       = lipgloss.NewStyle().Foreground(colCritters)
	helpStyle          = lipgloss.NewStyle().Foreground(colHelp)
	errorStyle         = lipgloss.NewStyle().Bold(true).Foreground(colErr)
)
