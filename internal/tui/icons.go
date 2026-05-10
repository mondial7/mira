package tui

import "github.com/charmbracelet/lipgloss"

// Each icon is exactly 3 lines tall and 5 columns wide so the grid layout
// math stays simple. The "Selected" variants use heavy box-drawing chars
// so the highlight is legible even when ANSI colors are stripped.
const (
	folderArt = `╭───╮
│ ▸ │
╰───╯`
	folderArtSelected = `┏━━━┓
┃ ▸ ┃
┗━━━┛`

	fileArt = `╭┄┄┄╮
┊ · ┊
╰┄┄┄╯`
	fileArtSelected = `┏━━━┓
┃ · ┃
┗━━━┛`

	linkArt = `╭───╮
│ ↪ │
╰───╯`
	linkArtSelected = `┏━━━┓
┃ ↪ ┃
┗━━━┛`

	parentArt = `╭───╮
│ ↑ │
╰───╯`
	parentArtSelected = `┏━━━┓
┃ ↑ ┃
┗━━━┛`
)

const (
	// CellWidth is the column width of one grid cell. CellHeight covers the
	// 3-line icon plus a 1-line label.
	CellWidth  = 12
	CellHeight = 4
)

var (
	colorAccent  = lipgloss.Color("69")  // soft purple-blue
	colorDim     = lipgloss.Color("245") // muted gray
	colorFolder  = lipgloss.Color("39")  // blue
	colorFile    = lipgloss.Color("252") // light gray
	colorLink    = lipgloss.Color("87")  // cyan
	colorErr     = lipgloss.Color("203") // red
	colorPathBg  = lipgloss.Color("236") // dark gray bg

	pathStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("231")).
			Background(colorPathBg).
			Padding(0, 1)

	labelStyle         = lipgloss.NewStyle().Foreground(colorFile)
	labelDirStyle      = lipgloss.NewStyle().Foreground(colorFolder)
	labelLinkStyle     = lipgloss.NewStyle().Foreground(colorLink)
	labelSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)

	helpStyle  = lipgloss.NewStyle().Foreground(colorDim)
	errorStyle = lipgloss.NewStyle().Bold(true).Foreground(colorErr)
)
