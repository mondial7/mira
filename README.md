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

↑↓←→/wasd move · ⏎ open · o launch · ⌫ up · h hidden · f find · . settings · e end here · q quit
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

### Homebrew (macOS / Linux)

```sh
brew install mondial7/tap/mira
```

The tap is published automatically on every release, so `brew upgrade
mira` always pulls the latest.

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
  --cd-file PATH  on 'e' (end here), write the chosen directory to PATH
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
| `o`                       | Launch highlighted file/folder in the OS default app |
| `Backspace` / `Esc`       | Go up one directory                           |
| `Home` / `g`              | Jump to first item                            |
| `End` / `G`               | Jump to last item                             |
| `h`                       | Toggle hidden (dotfile) entries               |
| `f`                       | Find — start a fuzzy search                   |
| `.`                       | Open the settings overlay                     |
| `q` / `Ctrl-C`            | Quit                                          |
| `e`                       | End here — quit and `cd` into the directory † |
| Mouse click on a folder   | Enter that folder                             |
| Mouse click on `..`       | Go up                                         |

† Requires the shell wrapper from the next section.

When the find bar is open: type to filter (case-insensitive subsequence
match), arrow keys to move within matches, `Enter` to open the
highlighted folder, `Esc` to cancel and restore the full listing.

### Settings

Press `.` in the file browser to open the settings overlay. It exposes
three knobs that change the look without leaving the TUI:

| Setting   | Values                          | Default |
| --------- | ------------------------------- | ------- |
| Theme     | `slate`, `forest`, `ocean`, `rose` | `slate` |
| Borders   | `fine`, `thick`, `dotted`       | `fine`  |
| Bionic    | `on`, `off`                     | `on`    |

Inside the overlay, `↑`/`↓` (or `w`/`s`) moves between rows, `←`/`→`
(or `a`/`d`) cycles the focused value, `Enter` cycles forward, and
`Esc` or `.` closes the overlay.

Choices persist between launches in a JSON file at
`os.UserConfigDir()/mira/config.json` — typically
`~/Library/Application Support/mira/config.json` on macOS,
`$XDG_CONFIG_HOME/mira/config.json` (or `~/.config/mira/config.json`)
on Linux, and `%AppData%\mira\config.json` on Windows. The file is
written only when you actually change a value; opening the overlay
and dismissing it without touching anything does not create one.
Delete the file to reset to defaults.

### End here & cd: shell wrapper

`e` ("end here") quits the TUI and asks the parent shell to `cd` into
whatever directory you ended up exploring. Plain `q` (lowercase)
quits without changing directory.

`e` only changes the parent shell's directory if you wrap the binary in
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

Now `m` opens the TUI; press `e` to end the session and have your shell
follow you into whatever directory you ended up exploring. Plain `q`
exits without changing the directory.

## Security

mira does not modify or transmit any file. The only side-effect it can
trigger is pressing `o`, which hands the highlighted path to the OS
default opener (`open` on macOS, `xdg-open` on Linux, `rundll32
url.dll,FileProtocolHandler` on Windows). The launched application
runs with your user's permissions — mira itself never reads or
executes the file's contents. See [SECURITY.md](SECURITY.md) for the
disclosure policy.

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
