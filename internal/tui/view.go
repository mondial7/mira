package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marcomondini/banana-four/internal/listing"
)

// View renders the current model state to a string for bubbletea to draw.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")
	b.WriteString(m.renderGrid())
	b.WriteString("\n")
	b.WriteString(m.renderFooter())
	return b.String()
}

func (m Model) renderHeader() string {
	path := m.cwd
	// Truncate long paths from the left so the leaf directory stays visible.
	maxPath := m.width - 4
	if maxPath > 0 && lipgloss.Width(path) > maxPath {
		path = "…" + path[lipgloss.Width(path)-maxPath+1:]
	}
	return pathStyle.Render("▸ " + path)
}

func (m Model) renderFooter() string {
	help := "↑↓←→ / hjkl move · ⏎ open · ⌫ up · click to enter · q quit"
	if m.err != nil {
		return errorStyle.Render("error: "+m.err.Error()) + "\n" + helpStyle.Render(help)
	}
	count := fmt.Sprintf("%d items", len(m.entries))
	return helpStyle.Render(count + " · " + help)
}

// renderGrid lays out the icon+label cells row by row. Icons are 3 lines
// tall, labels live on the 4th line; rows pack as tightly as the terminal
// width allows.
func (m Model) renderGrid() string {
	cols := m.cols()
	n := m.totalItems()
	if n == 0 {
		return helpStyle.Render("  (empty)")
	}

	var rows []string
	for start := 0; start < n; start += cols {
		end := start + cols
		if end > n {
			end = n
		}
		rows = append(rows, m.renderRow(start, end))
	}
	return strings.Join(rows, "\n")
}

// renderRow joins the icon-line + label-line stack for a horizontal slice
// of cells.
func (m Model) renderRow(start, end int) string {
	var (
		line0, line1, line2 strings.Builder
		labels              strings.Builder
	)
	for i := start; i < end; i++ {
		art := iconFor(m, i)
		artLines := strings.Split(art, "\n")
		line0.WriteString(centerInCell(artLines[0], CellWidth))
		line1.WriteString(centerInCell(artLines[1], CellWidth))
		line2.WriteString(centerInCell(artLines[2], CellWidth))
		labels.WriteString(centerInCell(labelFor(m, i), CellWidth))
	}
	return strings.Join([]string{
		line0.String(),
		line1.String(),
		line2.String(),
		labels.String(),
	}, "\n")
}

// iconFor renders the colorized ASCII art for the entry at index i.
func iconFor(m Model, i int) string {
	selected := i == m.cursor
	var art string
	var color lipgloss.Color

	switch {
	case m.isParent(i):
		art = pickArt(parentArt, parentArtSelected, selected)
		color = colorAccent
	default:
		e := m.entryAt(i)
		switch {
		case e.IsSymlink:
			art = pickArt(linkArt, linkArtSelected, selected)
			color = colorLink
		case e.IsDir:
			art = pickArt(folderArt, folderArtSelected, selected)
			color = colorFolder
		default:
			art = pickArt(fileArt, fileArtSelected, selected)
			color = colorFile
		}
	}

	if selected {
		color = colorAccent
	}
	return lipgloss.NewStyle().Foreground(color).Render(art)
}

func pickArt(unselected, selected string, isSelected bool) string {
	if isSelected {
		return selected
	}
	return unselected
}

// labelFor returns the colorized name shown beneath an icon, truncated to
// the cell width.
func labelFor(m Model, i int) string {
	var name string
	var style lipgloss.Style

	switch {
	case m.isParent(i):
		name = ".."
		style = labelStyle
	default:
		e := m.entryAt(i)
		name = e.Name
		switch {
		case e.IsDir:
			style = labelDirStyle
		case e.IsSymlink:
			style = labelLinkStyle
		default:
			style = labelStyle
		}
	}

	if i == m.cursor {
		style = labelSelectedStyle
	}

	name = truncate(name, CellWidth-2)
	return style.Render(name)
}

// centerInCell pads s with spaces so it occupies exactly width display
// columns. lipgloss.Width handles ANSI escape sequences correctly.
func centerInCell(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	left := (width - w) / 2
	right := width - w - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

// FlatList renders entries as a plain, unstyled list — used for piped output.
// Kept here so the formatting policy lives next to the interactive renderer.
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
