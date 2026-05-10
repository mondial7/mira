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
	cdMode := fs.Bool("cd", false, "force TUI mode; on Q-quit print the final directory to stdout (for shell-wrapper integration)")
	showVersion := fs.Bool("version", false, "print version and exit")

	fs.Usage = func() {
		fmt.Fprintf(stderr, `Usage: %s [options] [path]

A pretty, interactive folder visualizer.

By default banana-four opens a TUI. Pipe the output or pass --list to get
a plain listing instead. Pass --cd from a shell wrapper to capture the
final directory on a "Q" quit.

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
	case *cdMode:
		// Wrapper-driven invocation: stdout is captured by `$(...)`, the
		// TUI still runs on /dev/tty, and the chosen path goes to stdout.
		return runTUI(stdout, stderr, root, opts, true)
	case !isTTY(stdout):
		return runFlat(stdout, stderr, root, opts)
	default:
		return runTUI(stdout, stderr, root, opts, false)
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

func runTUI(stdout, stderr *os.File, root string, opts listing.Options, emitCWD bool) int {
	model, err := tui.New(root, opts)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	progOpts := []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
	// In --cd mode our stdout is captured by `$(banana-four --cd)`, so we
	// can't let bubbletea render into it (the wrapper would otherwise
	// `cd` into a string of escape codes). Route the TUI through
	// /dev/tty and keep stdout free for the chosen-directory print.
	if emitCWD {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			fmt.Fprintf(stderr, "error: --cd needs a controlling terminal: %v\n", err)
			return 1
		}
		defer tty.Close()
		progOpts = append(progOpts, tea.WithInput(tty), tea.WithOutput(tty))
	}

	prog := tea.NewProgram(model, progOpts...)
	final, err := prog.Run()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if emitCWD {
		if mf, ok := final.(tui.Model); ok && mf.QuitWithCD {
			fmt.Fprintln(stdout, mf.CWD())
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
