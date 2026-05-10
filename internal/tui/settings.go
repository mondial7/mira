package tui

import "github.com/charmbracelet/lipgloss"

// ColorTheme picks the active palette. Slate is the original look; the
// other three are mood variants.
type ColorTheme int

const (
	ThemeSlate  ColorTheme = iota // default — neutral grays + amber accent
	ThemeForest                   // greens + lime accent
	ThemeOcean                    // blues + cyan accent
	ThemeRose                     // warm pinks + rose accent
)

// allThemes is the cycle order used by the settings view. Keep Slate first
// so cycling forward from defaults walks through every variant before
// wrapping.
var allThemes = []ColorTheme{ThemeSlate, ThemeForest, ThemeOcean, ThemeRose}

// Label returns the human-readable name shown in the settings view.
func (t ColorTheme) Label() string {
	switch t {
	case ThemeForest:
		return "forest"
	case ThemeOcean:
		return "ocean"
	case ThemeRose:
		return "rose"
	default:
		return "slate"
	}
}

// BorderPreset chooses the glyph set used for card borders.
type BorderPreset int

const (
	BorderFine   BorderPreset = iota // default — rounded for dirs, double-line for files, heavy for selected
	BorderThick                      // heavy single-line for every card; selection signalled by colour
	BorderDotted                     // rounded corners with dotted edges
)

var allBorders = []BorderPreset{BorderFine, BorderThick, BorderDotted}

func (b BorderPreset) Label() string {
	switch b {
	case BorderThick:
		return "thick"
	case BorderDotted:
		return "dotted"
	default:
		return "fine"
	}
}

// Settings holds the user-tunable presentation knobs. Defaults match the
// pre-settings v0.1 look so existing users see no behavioural change until
// they open the settings view themselves.
type Settings struct {
	Theme   ColorTheme
	Borders BorderPreset
	Bionic  bool
}

// DefaultSettings returns the v0.1 baseline: slate palette, fine borders,
// bionic reading on.
func DefaultSettings() Settings {
	return Settings{
		Theme:   ThemeSlate,
		Borders: BorderFine,
		Bionic:  true,
	}
}

// palette holds the raw colour codes used to derive lipgloss styles.
// Each ColorTheme maps to one of these.
type palette struct {
	dim      lipgloss.Color // borders, hidden entries
	mid      lipgloss.Color // stats text
	name     lipgloss.Color // entry name
	nameSel  lipgloss.Color // selected entry name
	accent   lipgloss.Color // selection emphasis + scroll indicator
	critters lipgloss.Color // bottom-right cat & dog
	help     lipgloss.Color // footer help line
	pathBg   lipgloss.Color // path-bar background
	pathFg   lipgloss.Color // path-bar foreground
	err      lipgloss.Color // error text
}

func paletteFor(t ColorTheme) palette {
	switch t {
	case ThemeForest:
		return palette{
			dim:      lipgloss.Color("240"),
			mid:      lipgloss.Color("108"),
			name:     lipgloss.Color("151"),
			nameSel:  lipgloss.Color("230"),
			accent:   lipgloss.Color("119"),
			critters: lipgloss.Color("114"),
			help:     lipgloss.Color("242"),
			pathBg:   lipgloss.Color("22"),
			pathFg:   lipgloss.Color("230"),
			err:      lipgloss.Color("203"),
		}
	case ThemeOcean:
		return palette{
			dim:      lipgloss.Color("240"),
			mid:      lipgloss.Color("110"),
			name:     lipgloss.Color("153"),
			nameSel:  lipgloss.Color("231"),
			accent:   lipgloss.Color("39"),
			critters: lipgloss.Color("117"),
			help:     lipgloss.Color("242"),
			pathBg:   lipgloss.Color("24"),
			pathFg:   lipgloss.Color("231"),
			err:      lipgloss.Color("203"),
		}
	case ThemeRose:
		return palette{
			dim:      lipgloss.Color("240"),
			mid:      lipgloss.Color("180"),
			name:     lipgloss.Color("224"),
			nameSel:  lipgloss.Color("231"),
			accent:   lipgloss.Color("211"),
			critters: lipgloss.Color("217"),
			help:     lipgloss.Color("242"),
			pathBg:   lipgloss.Color("89"),
			pathFg:   lipgloss.Color("231"),
			err:      lipgloss.Color("203"),
		}
	default: // ThemeSlate
		return palette{
			dim:      lipgloss.Color("240"),
			mid:      lipgloss.Color("245"),
			name:     lipgloss.Color("250"),
			nameSel:  lipgloss.Color("231"),
			accent:   lipgloss.Color("214"),
			critters: lipgloss.Color("180"),
			help:     lipgloss.Color("242"),
			pathBg:   lipgloss.Color("236"),
			pathFg:   lipgloss.Color("231"),
			err:      lipgloss.Color("203"),
		}
	}
}

// themeStyles is the per-theme lipgloss style bundle the renderer uses.
// One instance is built whenever the active theme changes; everything else
// reads it through the model.
type themeStyles struct {
	path             lipgloss.Style
	border           lipgloss.Style
	selected         lipgloss.Style
	name             lipgloss.Style
	nameSelected     lipgloss.Style
	stats            lipgloss.Style
	statsSelected    lipgloss.Style
	nameHidden       lipgloss.Style
	statsHidden      lipgloss.Style
	critter          lipgloss.Style
	help             lipgloss.Style
	errorText        lipgloss.Style
	scrollIndicator  lipgloss.Style
	searchQuery      lipgloss.Style
	searchCursor     lipgloss.Style
	settingsTitle    lipgloss.Style
	settingsLabel    lipgloss.Style
	settingsValue    lipgloss.Style
	settingsCursor   lipgloss.Style
	settingsBorder   lipgloss.Style
	settingsInactive lipgloss.Style
}

func (p palette) styles() themeStyles {
	return themeStyles{
		path: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.pathFg).
			Background(p.pathBg).
			Padding(0, 1),
		border:           lipgloss.NewStyle().Foreground(p.dim),
		selected:         lipgloss.NewStyle().Foreground(p.accent).Bold(true),
		name:             lipgloss.NewStyle().Foreground(p.name),
		nameSelected:     lipgloss.NewStyle().Foreground(p.nameSel).Bold(true),
		stats:            lipgloss.NewStyle().Foreground(p.mid),
		statsSelected:    lipgloss.NewStyle().Foreground(p.mid).Bold(true),
		nameHidden:       lipgloss.NewStyle().Foreground(p.dim).Italic(true),
		statsHidden:      lipgloss.NewStyle().Foreground(p.dim).Italic(true),
		critter:          lipgloss.NewStyle().Foreground(p.critters),
		help:             lipgloss.NewStyle().Foreground(p.help),
		errorText:        lipgloss.NewStyle().Bold(true).Foreground(p.err),
		scrollIndicator:  lipgloss.NewStyle().Foreground(p.accent).Bold(true),
		searchQuery:      lipgloss.NewStyle().Foreground(p.nameSel).Bold(true),
		searchCursor:     lipgloss.NewStyle().Foreground(p.accent).Bold(true),
		settingsTitle:    lipgloss.NewStyle().Foreground(p.accent).Bold(true),
		settingsLabel:    lipgloss.NewStyle().Foreground(p.name),
		settingsValue:    lipgloss.NewStyle().Foreground(p.nameSel).Bold(true),
		settingsCursor:   lipgloss.NewStyle().Foreground(p.accent).Bold(true),
		settingsBorder:   lipgloss.NewStyle().Foreground(p.accent),
		settingsInactive: lipgloss.NewStyle().Foreground(p.help),
	}
}

// glyphSet is the six box-drawing characters that bound a card.
type glyphSet struct {
	tl, tr, bl, br, h, v string
}

var (
	glyphsRound  = glyphSet{tl: "╭", tr: "╮", bl: "╰", br: "╯", h: "─", v: "│"}
	glyphsDouble = glyphSet{tl: "╔", tr: "╗", bl: "╚", br: "╝", h: "═", v: "║"}
	glyphsHeavy  = glyphSet{tl: "┏", tr: "┓", bl: "┗", br: "┛", h: "━", v: "┃"}
	glyphsDotted = glyphSet{tl: "╭", tr: "╮", bl: "╰", br: "╯", h: "┈", v: "┊"}
)

// pickGlyphs returns the border glyph set for a single card given the
// active preset, the card's kind, and whether it's selected.
//
// BorderFine keeps the original v0.1 look (rounded for dirs, double-line
// for files, heavy for the selected card). BorderThick paints every card
// with heavy borders — the selection signal becomes colour, not shape.
// BorderDotted uses dotted edges across the board.
func pickGlyphs(preset BorderPreset, kind cardKind, selected bool) glyphSet {
	switch preset {
	case BorderThick:
		return glyphsHeavy
	case BorderDotted:
		return glyphsDotted
	default: // BorderFine
		switch {
		case selected:
			return glyphsHeavy
		case kind == kindFile:
			return glyphsDouble
		default:
			return glyphsRound
		}
	}
}

// settingsField identifies a row in the settings view.
type settingsField int

const (
	settingTheme settingsField = iota
	settingBorders
	settingBionic
)

// settingsFields is the row order shown in the settings view.
var settingsFields = []settingsField{settingTheme, settingBorders, settingBionic}

// cycleTheme advances the theme one step in either direction.
func (s *Settings) cycleTheme(delta int) {
	s.Theme = ColorTheme(cycleIndex(int(s.Theme), len(allThemes), delta))
}

// cycleBorders advances the border preset one step.
func (s *Settings) cycleBorders(delta int) {
	s.Borders = BorderPreset(cycleIndex(int(s.Borders), len(allBorders), delta))
}

// toggleBionic flips the bionic-reading flag.
func (s *Settings) toggleBionic() { s.Bionic = !s.Bionic }

// cycleIndex wraps i+delta into [0,n). Negative deltas walk backwards.
func cycleIndex(i, n, delta int) int {
	if n <= 0 {
		return 0
	}
	out := (i + delta) % n
	if out < 0 {
		out += n
	}
	return out
}
