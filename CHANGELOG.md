# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Interactive TUI file browser built on
  [`bubbletea`](https://github.com/charmbracelet/bubbletea) with mouse +
  keyboard navigation, single-layer grid view of cards.
- Bigger cards (20×6 chars) with the entry name and a stat line —
  child-count for directories, human-readable size for files, target for
  symlinks — drawn inside each card.
- Distinct selection cues that survive even with colors stripped:
  selected cards swap to heavy box-drawing borders and a bold-glyph
  selection symbol (▶ vs ▸, ◆ vs ·, ⇒ vs ↪, ▲ vs ↑).
- Monochromatic slate-gray palette with a single warm-amber accent for
  selection.
- Bottom-right ASCII critters (cat + dog) that track the cursor's
  horizontal position and run an idle blink/wag animation on a
  600 ms tick.
- `.gitignore`-aware listing (recursive across nested `.gitignore` files
  up to the git root); toggle with `--no-ignore`.
- Pipe-friendly flat-list mode auto-enabled when stdout is not a TTY,
  forced via `--list`.
- Hidden file toggle (`-a`) and dirs-only filter (`-d`).
- Pre-built binaries for darwin/linux/windows on amd64 and arm64,
  published by GoReleaser with checksums + SBOM.
