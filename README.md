# mira

[![ci](https://github.com/mondial7/mira/actions/workflows/ci.yml/badge.svg)](https://github.com/mondial7/mira/actions/workflows/ci.yml)
[![codeql](https://github.com/mondial7/mira/actions/workflows/codeql.yml/badge.svg)](https://github.com/mondial7/mira/actions/workflows/codeql.yml)
[![release](https://img.shields.io/github/v/release/mondial7/mira?display_name=tag&sort=semver)](https://github.com/mondial7/mira/releases/latest)
[![go reference](https://pkg.go.dev/badge/github.com/mondial7/mira.svg)](https://pkg.go.dev/github.com/mondial7/mira)
[![go report](https://goreportcard.com/badge/github.com/mondial7/mira)](https://goreportcard.com/report/github.com/mondial7/mira)
[![license](https://img.shields.io/github/license/mondial7/mira)](LICENSE)

A pretty, interactive folder visualizer for the terminal — a desktop-style
file browser with ASCII-art icons, mouse clicks, keyboard navigation, and
`.gitignore`-aware listing.

```
 ▸ /Users/you/code/some-repo

╭──────────────────╮  ┏━━━━━━━━━━━━━━━━━━┓  ╭──────────────────╮  ╭┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈╮
│ ↑  ..            │  ┃ ▶  docs          ┃  │ ▸  internal      │  ┊ ·  README.md     ┊
│                  │  ┃                  ┃  │                  │  ┊                  ┊
│   go up          │  ┃   3 items        ┃  │   2 items        │  ┊   4.3KB          ┊
│                  │  ┃                  ┃  │                  │  ┊                  ┊
╰──────────────────╯  ┗━━━━━━━━━━━━━━━━━━┛  ╰──────────────────╯  ╰┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈╯

                                                 /\_/\     /^.^\
                                                ( o.o )   ( o.o )
                                                 > ^ <     v=-=v

↑↓←→/wasd move · ⏎ open · ⌫ up · h hidden · f find · Q cd here · q quit
```

## Features

- **Interactive grid view** — every directory is a single layer of clickable
  cards. No deep tree to scan with your eyes.
- **Real mouse support** — click any folder to enter it, click `..` to go up.
- **Keyboard-first** — arrow keys or `wasd`, `Enter` to open, `Backspace` /
  `Esc` to go up, `q` to quit.
- **Gitignore-aware** — files matched by your project's `.gitignore` are
  hidden by default, just like in your editor. Pass `--no-ignore` to see
  everything.
- **Pipe-friendly** — when stdout is not a terminal (or you pass `--list`),
  output is a clean, plain listing you can pipe into `grep`, `xargs`, etc.
- **Zero runtime config** — single binary, no shell integration required.

## Install

### Pre-built binaries

Download the binary for your platform from the
[latest release](https://github.com/mondial7/mira/releases/latest)
and put it on your `PATH`. Each release includes SHA-256 checksums and an
SBOM.

### From source

```sh
go install github.com/mondial7/mira@latest
```

## Usage

```text
Usage: mira [options] [path]

Options:
  -a              show hidden (dotfile) entries
  -d              list directories only
  --list          force flat-list output instead of the TUI
  --no-ignore     disable .gitignore filtering
  --cd-file PATH  on Q-quit, write the chosen directory to PATH
                  (used by the shell wrapper — see below)
  --version       print version and exit
```

### Examples

Browse the current directory interactively:

```sh
mira
```

Browse a specific path, including hidden files and gitignored entries:

```sh
mira -a --no-ignore ~/code/some-repo
```

Pipe a flat listing into your shell:

```sh
mira --list | grep .go
```

### Keybindings

| Key                       | Action                                        |
| ------------------------- | --------------------------------------------- |
| `←` / `→` / `a` / `d`     | Move cursor left/right                        |
| `↑` / `↓` / `w` / `s`     | Move cursor up/down                           |
| `Enter` / `Space`         | Open selected folder                          |
| `Backspace` / `Esc`       | Go up one directory                           |
| `Home` / `g`              | Jump to first item                            |
| `End` / `G`               | Jump to last item                             |
| `h`                       | Toggle hidden (dotfile) entries               |
| `f`                       | Find — start a fuzzy search                   |
| `q` / `Ctrl-C`            | Quit                                          |
| `Q`                       | Quit and `cd` into the explored directory †   |
| Mouse click on a folder   | Enter that folder                             |
| Mouse click on `..`       | Go up                                         |

† Requires the shell wrapper from the next section.

When the find bar is open: type to filter (case-insensitive subsequence
match), arrow keys to move within matches, `Enter` to open the
highlighted folder, `Esc` to cancel and restore the full listing.

### Quit & cd: shell wrapper

> **Status (v0.1):** the `Q` quit-and-`cd` flow is **known to be broken**
> in the current release — under some shells/terminals the parent shell
> does not follow into the chosen directory even with the wrapper below.
> Tracked for a fix before **v1.0**. Plain `q` (lowercase) quits cleanly
> and is unaffected.

`Q` only changes the parent shell's directory if you wrap the binary in
a shell function. The wrapper passes a temp file via `--cd-file` and
reads it back after the TUI exits — this is bulletproof against any
output that the terminal-restore step might emit.

Drop one of these in your shell rc file (rename `m` to whatever you like):

```sh
# ~/.bashrc / ~/.zshrc
m() {
  local f
  f=$(mktemp -t mira.XXXXXX) || return
  command mira --cd-file "$f" "$@"
  if [ -s "$f" ]; then
    cd -- "$(cat "$f")"
  fi
  rm -f "$f"
}
```

```fish
# ~/.config/fish/functions/m.fish
function m
  set f (mktemp -t mira.XXXXXX); or return
  command mira --cd-file $f $argv
  if test -s $f
    cd -- (cat $f)
  end
  rm -f $f
end
```

Now `m` opens the TUI; press `Q` to leave and have your shell follow you
into whatever directory you ended up exploring. Plain `q` exits without
changing the directory.

## Roadmap to v1.0

`v0.1` is the first public release — usable end-to-end, but a few items
are explicitly deferred until **v1.0**:

- **Fix `Q` quit-and-`cd` handoff.** See the known issue noted in the
  shell-wrapper section above.
- **Settings view + customisation pattern.** A first-class `,` (or
  similar) settings screen will land before v1, together with the
  pattern that future customisation surfaces (theme, keymap, default
  flags) will follow. Until then the TUI is intentionally
  zero-configuration.
- **Homebrew tap.** Goreleaser support is wired but commented out;
  enabling it on a `homebrew-tap` repo is on the v1 checklist so macOS
  users can `brew install mondial7/tap/mira`.

Smaller polish items are tracked in
[GitHub issues](https://github.com/mondial7/mira/issues).

## Security

This tool is **read-only**. It lists directory contents; it does not open,
modify, execute, or transmit any file. See [SECURITY.md](SECURITY.md) for
the disclosure policy.

## Contributing

Pull requests are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the
full guide. The TL;DR:

```sh
git clone https://github.com/mondial7/mira
cd mira
make test
go run .
```

A `Makefile` wraps the common dev loop: `make test`, `make lint`,
`make build`, `make release-snapshot`. Please run `make lint` before
opening a PR.

## Project structure

```
mira/
├── main.go                  entry point: flags + TTY detection
├── internal/
│   ├── listing/             pure logic: read + filter + sort + gitignore
│   └── tui/                 bubbletea Model/View/Update + ASCII art
├── docs/                    GitHub Pages landing site
└── .github/                 CI, security scans, issue templates, release
```

## Acknowledgements

Built on the excellent [`bubbletea`](https://github.com/charmbracelet/bubbletea)
and [`lipgloss`](https://github.com/charmbracelet/lipgloss) by Charm, plus
[`go-gitignore`](https://github.com/sabhiram/go-gitignore) by Shaba Abhiram.

## License

[MIT](LICENSE)
