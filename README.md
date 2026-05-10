# banana-four

A pretty, interactive folder visualizer for the terminal — a desktop-style
file browser with ASCII-art icons, mouse clicks, keyboard navigation, and
`.gitignore`-aware listing.

```
 ▸ /Users/you/code/banana-four

╭──────────────────╮  ┏━━━━━━━━━━━━━━━━━━┓  ╭──────────────────╮  ╭┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈╮
│ ↑  ..            │  ┃ ▶  docs          ┃  │ ▸  internal      │  ┊ ·  README.md     ┊
│                  │  ┃                  ┃  │                  │  ┊                  ┊
│   go up          │  ┃   3 items        ┃  │   2 items        │  ┊   4.3KB          ┊
│                  │  ┃                  ┃  │                  │  ┊                  ┊
╰──────────────────╯  ┗━━━━━━━━━━━━━━━━━━┛  ╰──────────────────╯  ╰┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈╯

                                                 /\_/\     /^.^\
                                                ( o.o )   ( o.o )
                                                 > ^ <     v=-=v

5 items · ↑↓←→/wasd move · ⏎ open · ⌫ up · h hidden · Q cd here · q quit
```

## Features

- **Interactive grid view** — every directory is a single layer of clickable
  cards. No deep tree to scan with your eyes.
- **Real mouse support** — click any folder to enter it, click `..` to go up.
- **Keyboard-first** — arrow keys or `hjkl`, `Enter` to open, `Backspace` /
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
[latest release](https://github.com/mondial7/banana-four/releases/latest)
and put it on your `PATH`. Each release includes SHA-256 checksums and an
SBOM.

### From source

```sh
go install github.com/mondial7/banana-four@latest
```

## Usage

```text
Usage: banana-four [options] [path]

Options:
  -L              max display depth (legacy tree mode — see --list)
  -a              show hidden (dotfile) entries
  -d              list directories only
  --list          force flat-list output instead of the TUI
  --no-color      disable ANSI color output
  --no-ignore     disable .gitignore filtering
  --version       print version and exit
```

### Examples

Browse the current directory interactively:

```sh
banana-four
```

Browse a specific path, including hidden files and gitignored entries:

```sh
banana-four -a --no-ignore ~/code/some-repo
```

Pipe a flat listing into your shell:

```sh
banana-four --list | grep .go
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

`Q` only changes the parent shell's directory if you wrap the binary in
a shell function. The wrapper passes a temp file via `--cd-file` and
reads it back after the TUI exits — this is bulletproof against any
output that the terminal-restore step might emit.

Drop one of these in your shell rc file:

```sh
# ~/.bashrc / ~/.zshrc
bf() {
  local f
  f=$(mktemp -t bf.XXXXXX) || return
  command banana-four --cd-file "$f" "$@"
  if [ -s "$f" ]; then
    cd -- "$(cat "$f")"
  fi
  rm -f "$f"
}
```

```fish
# ~/.config/fish/functions/bf.fish
function bf
  set f (mktemp -t bf.XXXXXX); or return
  command banana-four --cd-file $f $argv
  if test -s $f
    cd -- (cat $f)
  end
  rm -f $f
end
```

Now `bf` opens the TUI; press `Q` to leave and have your shell follow you
into whatever directory you ended up exploring. Plain `q` exits without
changing the directory.

## Security

This tool is **read-only**. It lists directory contents; it does not open,
modify, execute, or transmit any file. See [SECURITY.md](SECURITY.md) for
the disclosure policy.

## Contributing

Pull requests are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the
full guide. The TL;DR:

```sh
git clone https://github.com/mondial7/banana-four
cd banana-four
go test ./...
go run .
```

Please run `go vet ./...` and `gofmt -s -w .` before opening a PR.

## Project structure

```
banana-four/
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
