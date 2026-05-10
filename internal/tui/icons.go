package tui

// Card geometry. Cells are 6 lines tall (top border, name, separator,
// stats, blank, bottom border) plus one blank line of vertical padding
// between rows. The math elsewhere uses CellHeight (the card itself) and
// rowGap (the spacer) so click-to-cell mapping stays consistent.
const (
	// MinCellWidth is the floor for dynamic cell sizing. Cards never get
	// narrower than this even when entries are tiny, so the UI keeps a
	// consistent feel.
	MinCellWidth = 20
	CellHeight   = 6
	rowGap       = 1 // blank line drawn between rows of cards
	colGap       = 2 // blank columns drawn between cards in a row
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

// defaultStyles is the slate-theme style bundle used when a Model has no
// settings of its own (zero-value Model in tests, etc.) and as the source
// for the package-level style aliases below.
var defaultStyles = paletteFor(ThemeSlate).styles()

// Package-level aliases preserved for tests and ad-hoc callers. Active
// rendering goes through Model.styles so theme changes take effect at
// runtime.
var (
	pathStyle            = defaultStyles.path
	borderStyle          = defaultStyles.border
	selectedStyle        = defaultStyles.selected
	nameStyle            = defaultStyles.name
	nameSelectedStyle    = defaultStyles.nameSelected
	statsStyle           = defaultStyles.stats
	statsSelectedStyle   = defaultStyles.statsSelected
	nameHiddenStyle      = defaultStyles.nameHidden
	statsHiddenStyle     = defaultStyles.statsHidden
	critterStyle         = defaultStyles.critter
	helpStyle            = defaultStyles.help
	errorStyle           = defaultStyles.errorText
	scrollIndicatorStyle = defaultStyles.scrollIndicator
	searchQueryStyle     = defaultStyles.searchQuery
	searchCursorStyle    = defaultStyles.searchCursor
)
