# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Initial release: interactive TUI file browser built on
  [`bubbletea`](https://github.com/charmbracelet/bubbletea) with mouse +
  keyboard navigation.
- ASCII-art folder/file/symlink icons with selected-state styling.
- `.gitignore`-aware listing (recursive across nested `.gitignore` files
  up to the git root); toggle with `--no-ignore`.
- Pipe-friendly flat-list mode auto-enabled when stdout is not a TTY,
  forced via `--list`.
- Hidden file toggle (`-a`) and dirs-only filter (`-d`).
- Pre-built binaries for darwin/linux/windows on amd64 and arm64,
  published by GoReleaser with checksums + SBOM.
- Homebrew tap support.
