// Command banana-four is an interactive terminal file browser. By default
// it launches a desktop-style TUI with ASCII-art folder icons and mouse +
// keyboard navigation. When stdout is not a TTY (e.g. piping into another
// command) it prints a plain, gitignore-aware listing of the current
// directory instead, so it composes well in scripts.
//
// Pass --cd when invoking from a wrapper shell function: that forces the
// TUI on /dev/tty even when stdout is captured, and on a "Q" quit the
// chosen directory is printed to stdout so the wrapper can `cd` to it.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/banana-four/internal/listing"
	"github.com/mondial7/banana-four/internal/tui"
)

// version is overridden at build time via -ldflags by goreleaser.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet("banana-four", flag.ContinueOnError)
	fs.SetOutput(stderr)

	all := fs.Bool("a", false, "show hidden (dotfile) entries")
	dirs := fs.Bool("d", false, "list directories only")
	noIgnore := fs.Bool("no-ignore", false, "disable .gitignore filtering")
	listMode := fs.Bool("list", false, "force flat-list output instead of the TUI")
	cdFile := fs.String("cd-file", "", "write the chosen directory to PATH on Q-quit (for shell-wrapper integration)")
	showVersion := fs.Bool("version", false, "print version and exit")

	fs.Usage = func() {
		fmt.Fprintf(stderr, `Usage: %s [options] [path]

A pretty, interactive folder visualizer.

By default banana-four opens a TUI. Pipe the output or pass --list to get
a plain listing instead. Pass --cd-file from a shell wrapper to capture
the final directory on a "Q" quit.

Options:
`, filepath.Base(os.Args[0]))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	root := "."
	if fs.NArg() > 0 {
		root = fs.Arg(0)
	}

	opts := listing.Options{
		ShowHidden:   *all,
		DirsOnly:     *dirs,
		UseGitignore: !*noIgnore,
	}

	switch {
	case *listMode:
		return runFlat(stdout, stderr, root, opts)
	case *cdFile != "":
		// Wrapper-driven invocation: TUI runs as normal on the inherited
		// stdout/stdin, and the chosen path is written to the file on Q.
		return runTUI(stderr, root, opts, *cdFile)
	case !isTTY(stdout):
		return runFlat(stdout, stderr, root, opts)
	default:
		return runTUI(stderr, root, opts, "")
	}
}

func runFlat(stdout, stderr *os.File, root string, opts listing.Options) int {
	entries, err := listing.List(root, opts)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprint(stdout, tui.FlatList(entries))
	return 0
}

func runTUI(stderr *os.File, root string, opts listing.Options, cdFile string) int {
	model, err := tui.New(root, opts)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	prog := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	final, err := prog.Run()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if cdFile != "" {
		if mf, ok := final.(tui.Model); ok && mf.QuitWithCD {
			// File-based handoff: bulletproof against any stdout interaction
			// from bubbletea's altscreen teardown. The wrapper reads the file.
			if err := os.WriteFile(cdFile, []byte(mf.CWD()+"\n"), 0o644); err != nil {
				fmt.Fprintf(stderr, "warning: could not write cd-file: %v\n", err)
			}
		}
	}
	return 0
}

// isTTY reports whether f is connected to a terminal. When false, we fall
// back to the flat-list mode so output composes cleanly with shell pipes.
func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
