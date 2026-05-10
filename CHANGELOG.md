# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- **`Q` → `e` for "end here".** The capital-Q quit-and-`cd` binding
  has been retired in favour of lowercase `e`. The v0.1 `Q` flow had
  a stdout-handoff bug under some shells/terminals; renaming it
  alongside the fix avoids confusion with the broken keystroke and
  reads better next to the existing `q` quit.

### Added

- **Settings overlay (`.`).** First-class settings screen with three
  knobs: colour theme (`slate` / `forest` / `ocean` / `rose`), border
  preset (`fine` / `thick` / `dotted`), and a bionic-reading on/off
  toggle. Defaults match the v0.1 look. Settings are session-scoped;
  a persisted-config story is the remaining v1 piece.

### Planned for v1.0

- Persisted settings (the `.` overlay survives across launches).
- Optional Homebrew tap (`brew install mondial7/tap/mira`).

## [0.1.0] - 2026-05-10

First public release of `mira` (the project was previously codenamed
`banana-four` while it was private; nothing under that name was ever
published). Everything below previously sat under "Unreleased".

### Changed

- Letter-based navigation moved from `hjkl` (vim-style) to `adws`
  (gaming-style). `h` is now reserved for the hidden-toggle.
- Quit-and-cd now uses `--cd-file PATH` instead of `--cd` for a
  bulletproof file-based handoff (avoids stdout-capture fragility).
  The shell wrapper accordingly switched to `mktemp` + `cat`.

### Added

- `f` opens a fuzzy-search find bar replacing the summary line.
  Case-insensitive subsequence match, live-filtered cards, esc cancels,
  enter opens the highlighted match.
- Bionic Reading on entry names: the leading half of each
  word-segment (split on `_ - . space /`) is bolded so the eye can
  pattern-match faster. Skipped on `..` and selected entries.
- Viewport scrolling: tall listings no longer clip the top — content
  auto-scrolls to keep the cursor visible, with ▲/▼ amber indicators
  in the path bar's top-right when content extends beyond the viewport.

- `h` toggles dotfile visibility at runtime; hidden entries render in a
  dimmed-italic style so they're visible but obviously secondary.
- `Q` (capital Q) quits the TUI and prints the current directory to
  stdout; combined with the shell wrapper documented in the README, the
  parent shell follows you into whatever folder you ended up exploring.
- `--cd` flag forces TUI mode even when stdout is captured, so the
  shell-wrapper integration works through `$(...)`.

### Earlier in unreleased

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

[Unreleased]: https://github.com/mondial7/mira/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mondial7/mira/releases/tag/v0.1.0
