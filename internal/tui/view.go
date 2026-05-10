package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mondial7/banana-four/internal/listing"
)

// View renders the current model state. The structure is:
//
//	▸ /current/path                          ← path bar (1 line)
//	  3 folders · 8 files · ~4.2MB           ← stats summary (1 line)
//	                                         ← blank
//	┌── grid of cards, CellHeight tall ──┐
//	│  with rowGap blank line between    │
//	└────────────────────────────────────┘
//	                                         ← blank gap to footer
//	(critter strip — CritterHeight lines, right-aligned)
//	                                         ← blank
//	N items · ↑↓←→ ... q quit                ← help line
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteByte('\n')
	b.WriteString(m.renderSummary())
	b.WriteString("\n\n")
	b.WriteString(m.renderGrid())
	b.WriteString("\n\n")
	b.WriteString(renderCritters(m.viewWidth(), m.animFrame, m.cursorLookDir()))
	b.WriteString("\n\n")
	b.WriteString(m.renderFooter())
	return b.String()
}

// viewWidth returns at least 1 column so layout math always succeeds.
func (m Model) viewWidth() int {
	if m.width < 1 {
		return 1
	}
	return m.width
}

func (m Model) renderHeader() string {
	path := m.cwd
	indicator := m.scrollIndicator()
	indW := lipgloss.Width(indicator)

	// Reserve room for the scroll indicator at the right edge so the path
	// isn't pushed under it. The +2 is a one-space margin between the
	// path bar's right edge and the indicator.
	reserve := 0
	if indW > 0 {
		reserve = indW + 2
	}
	maxPath := m.viewWidth() - 4 - reserve
	if maxPath > 0 && lipgloss.Width(path) > maxPath {
		runes := []rune(path)
		if len(runes) > maxPath-1 {
			path = "…" + string(runes[len(runes)-(maxPath-1):])
		}
	}
	pathBar := pathStyle.Render("▸ " + path)

	if indW == 0 {
		return pathBar
	}

	pathBarW := lipgloss.Width(pathBar)
	spaces := m.viewWidth() - pathBarW - indW
	if spaces < 1 {
		spaces = 1
	}
	return pathBar + strings.Repeat(" ", spaces) + scrollIndicatorStyle.Render(indicator)
}

// scrollIndicator returns a 2-character glyph string showing whether the
// grid extends above or below the current viewport. Empty when the
// listing fits entirely.
func (m Model) scrollIndicator() string {
	canUp := m.scrollOffset > 0
	canDown := m.scrollOffset+m.visibleGridRows() < m.totalGridRows()
	if !canUp && !canDown {
		return ""
	}
	up := " "
	down := " "
	if canUp {
		up = "▲"
	}
	if canDown {
		down = "▼"
	}
	return up + down
}

// renderSummary draws the second header line with a quick recap of the
// current directory: how many folders + files it holds and their total
// size. The size is prefixed with "~" when the recursive walk hit its
// budget cap and so the number is approximate.
func (m Model) renderSummary() string {
	folders := plural(m.totalDirs, "folder", "folders")
	files := plural(m.totalFiles, "file", "files")

	size := listing.HumanSize(m.totalSize)
	if !m.sizeExact {
		size = "~" + size
	}
	leaf := filepath.Base(m.cwd)
	parts := []string{leaf, folders, files, size}
	return helpStyle.Render("  " + strings.Join(parts, " · "))
}

func (m Model) renderFooter() string {
	help := "↑↓←→/wasd move · ⏎ open · ⌫ up · h hidden · Q cd here · q quit"
	if m.err != nil {
		return errorStyle.Render("error: "+m.err.Error()) + "\n" + helpStyle.Render(help)
	}
	count := fmt.Sprintf("%d items", len(m.entries))
	return helpStyle.Render(count + " · " + help)
}

// renderGrid lays out the rows currently inside the scroll viewport,
// separating them with a blank line for breathing room. Off-screen rows
// are signalled by the ▲/▼ indicators in the header.
func (m Model) renderGrid() string {
	cols := m.cols()
	n := m.totalItems()
	if n == 0 {
		return helpStyle.Render("  (empty)")
	}

	visible := m.visibleGridRows()
	startRow := m.scrollOffset
	totalRows := m.totalGridRows()
	endRow := startRow + visible
	if endRow > totalRows {
		endRow = totalRows
	}

	var rows []string
	for r := startRow; r < endRow; r++ {
		start := r * cols
		end := start + cols
		if end > n {
			end = n
		}
		rows = append(rows, m.renderRow(start, end))
	}
	gap := "\n" + strings.Repeat("\n", rowGap)
	return strings.Join(rows, gap)
}

// renderRow renders one horizontal slice of cards by interleaving each
// card's lines, with a colGap-sized blank gutter between cards.
func (m Model) renderRow(start, end int) string {
	cards := make([][]string, end-start)
	for i := start; i < end; i++ {
		cards[i-start] = renderCard(m, i)
	}
	gutter := strings.Repeat(" ", colGap)
	var b strings.Builder
	for line := 0; line < CellHeight; line++ {
		for ci, card := range cards {
			if ci > 0 {
				b.WriteString(gutter)
			}
			b.WriteString(card[line])
		}
		if line < CellHeight-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// renderCard returns CellHeight lines of the card at index i. Each line is
// padded to exactly m.cellWidth() display columns so renderRow can
// concatenate horizontally without extra alignment.
func renderCard(m Model, i int) []string {
	selected := i == m.cursor
	parent := m.isParent(i)
	sym, name, stats, kind := cardContent(m, i)
	hidden := !parent && strings.HasPrefix(name, ".")

	tl, tr, bl, br, h, v, bs, ns, ss := cardChrome(selected, kind, hidden)
	cellW := m.cellWidth()
	innerWidth := cellW - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Reserve " sym  " (4 cols) so the name fits beside the symbol.
	maxNameDisplay := innerWidth - 4
	if maxNameDisplay < 1 {
		maxNameDisplay = 1
	}
	visibleName := clampPlain(name, maxNameDisplay)

	// Bionic reading: bold the leading half of each word-segment so the
	// eye can pattern-match faster. Skipped for the ".." entry and for
	// the selected entry (its style is already fully bold-accent — bionic
	// would have no contrast).
	useBionic := !parent && !selected
	styledName := styleName(visibleName, ns, useBionic)

	nameLine := " " + ns.Render(sym) + "  " + styledName
	statsLine := ss.Render("   " + clampPlain(stats, innerWidth-3))

	nameLine = padDisplayWidth(nameLine, innerWidth)
	statsLine = padDisplayWidth(statsLine, innerWidth)

	lines := make([]string, CellHeight)
	lines[lineTop] = bs.Render(tl + strings.Repeat(h, innerWidth) + tr)
	lines[lineName] = bs.Render(v) + nameLine + bs.Render(v)
	lines[lineSep] = bs.Render(v) + strings.Repeat(" ", innerWidth) + bs.Render(v)
	lines[lineStats] = bs.Render(v) + statsLine + bs.Render(v)
	lines[lineSpacer] = bs.Render(v) + strings.Repeat(" ", innerWidth) + bs.Render(v)
	lines[lineBottom] = bs.Render(bl + strings.Repeat(h, innerWidth) + br)
	return lines
}

// styleName renders name through base, optionally applying bionic-style
// bolding: each word-segment (split on _ - . space /) gets its leading
// half drawn with Bold added to the base style. Cap of 4 prevents very
// long words from being almost entirely bold.
func styleName(name string, base lipgloss.Style, bionic bool) string {
	if !bionic || name == "" {
		return base.Render(name)
	}
	bold := base.Bold(true)
	var b strings.Builder
	word := make([]rune, 0, 8)
	flush := func() {
		if len(word) == 0 {
			return
		}
		n := (len(word) + 1) / 2
		if n > 4 {
			n = 4
		}
		b.WriteString(bold.Render(string(word[:n])))
		if n < len(word) {
			b.WriteString(base.Render(string(word[n:])))
		}
		word = word[:0]
	}
	for _, r := range name {
		switch r {
		case '_', '-', '.', ' ', '/':
			flush()
			b.WriteString(base.Render(string(r)))
		default:
			word = append(word, r)
		}
	}
	flush()
	return b.String()
}

// padDisplayWidth right-pads s with spaces so its display width is
// exactly width columns. Unlike padToWidth it does not re-apply a style,
// which would corrupt strings that already contain ANSI escapes.
func padDisplayWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// clampPlain truncates a plain (un-styled) string to at most max display
// columns, appending an ellipsis when it had to cut.
func clampPlain(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > max {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// cardContent assembles the human-readable pieces shown inside the card:
// selection symbol, display name, single-line stats, and the entry kind
// used to pick border style. Names are returned untruncated; the renderer
// applies a final safety-net clamp only when no dynamic width can fit.
func cardContent(m Model, i int) (sym, name, stats string, kind cardKind) {
	if m.isParent(i) {
		sym = symParent
		if i == m.cursor {
			sym = symParentSelected
		}
		return sym, "..", "go up", kindParent
	}

	e := m.entryAt(i)
	selected := i == m.cursor
	switch {
	case e.IsSymlink:
		sym = symLink
		if selected {
			sym = symLinkSelected
		}
		kind = kindLink
		stats = "→ " + e.Target
	case e.IsDir:
		sym = symFolder
		if selected {
			sym = symFolderSelected
		}
		kind = kindDir
		stats = dirStats(e)
	default:
		sym = symFile
		if selected {
			sym = symFileSelected
		}
		kind = kindFile
		stats = listing.HumanSize(e.Size)
	}
	name = e.Name
	return
}

// cardTextLines returns the rendered display widths of the name and stats
// lines for entry i, BEFORE styling. computeCellWidth uses these to size
// cards.
func cardTextLines(m Model, i int) (nameW, statsW int) {
	sym, name, stats, _ := cardContent(m, i)
	nameLine := fmt.Sprintf(" %s  %s", sym, name)
	statsLine := fmt.Sprintf("   %s", stats)
	return lipgloss.Width(nameLine), lipgloss.Width(statsLine)
}

// dirStats formats a directory's stats line: child count + total size.
// Approximate sizes are flagged with a "~" prefix.
func dirStats(e listing.Entry) string {
	count := childCountLabel(e.ChildCount)
	size := listing.HumanSize(e.Size)
	if !e.SizeExact {
		size = "~" + size
	}
	return count + " · " + size
}

type cardKind int

const (
	kindParent cardKind = iota
	kindDir
	kindFile
	kindLink
)

// cardChrome returns the border glyphs + lipgloss styles for a card based
// on whether it's selected, what kind of entry it represents, and whether
// it's a hidden (dotfile) entry. Hidden entries inherit the kind's border
// style but get dimmed-italic text so they're visually secondary.
func cardChrome(selected bool, kind cardKind, hidden bool) (
	tl, tr, bl, br, h, v string,
	border, name, stats lipgloss.Style,
) {
	switch {
	case selected:
		tl, tr, bl, br = heavyTL, heavyTR, heavyBL, heavyBR
		h, v = heavyH, heavyV
		border = selectedStyle
		name = nameSelectedStyle
		stats = statsSelectedStyle
		return
	case kind == kindFile:
		tl, tr, bl, br = doubleTL, doubleTR, doubleBL, doubleBR
		h, v = doubleH, doubleV
		border = borderStyle
		name = nameStyle
		stats = statsStyle
	default:
		tl, tr, bl, br = roundTL, roundTR, roundBL, roundBR
		h, v = roundH, roundV
		border = borderStyle
		name = nameStyle
		stats = statsStyle
	}
	if hidden {
		name = nameHiddenStyle
		stats = statsHiddenStyle
	}
	return
}

// padToWidth applies the given style to s, then right-pads with spaces so
// the rendered display width is exactly width columns. lipgloss.Width
// understands ANSI escape sequences so the math stays correct.
func padToWidth(s string, width int, style lipgloss.Style) string {
	rendered := style.Render(s)
	w := lipgloss.Width(rendered)
	if w >= width {
		return rendered
	}
	return rendered + strings.Repeat(" ", width-w)
}

// clampDisplay truncates s with an ellipsis if its display width exceeds
// max columns. This is a safety net used only when the dynamic cell
// sizing couldn't grow large enough (e.g. in extremely narrow terminals).
func clampDisplay(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > max {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// childCountLabel renders the dir-stats summary with singular/plural and a
// graceful fallback when the count is unknown.
func childCountLabel(n int) string {
	switch {
	case n < 0:
		return "—"
	case n == 0:
		return "empty"
	case n == 1:
		return "1 item"
	default:
		return fmt.Sprintf("%d items", n)
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// FlatList renders entries as a plain, unstyled list — used for piped output.
func FlatList(entries []listing.Entry) string {
	var b strings.Builder
	for _, e := range entries {
		switch {
		case e.IsDir:
			b.WriteString(e.Name + "/\n")
		case e.IsSymlink:
			b.WriteString(e.Name + " -> " + e.Target + "\n")
		default:
			b.WriteString(e.Name + "\n")
		}
	}
	return b.String()
}
