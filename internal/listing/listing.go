// Package listing reads directory entries with optional .gitignore-aware
// filtering. It is pure logic: it does not render, take input, or touch
// terminal state, so it can be unit-tested without a TTY.
package listing

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry describes a single directory entry exposed to the UI layer.
type Entry struct {
	Name      string
	Path      string
	IsDir     bool
	IsSymlink bool
	Mode      os.FileMode
	Size      int64
	Target    string // resolved symlink target, empty for non-symlinks
	// ChildCount is the raw number of entries in this directory (or -1 if
	// not applicable / not readable). It includes dotfiles and ignored
	// entries on purpose: it's a quick "how full is this?" hint, not a
	// preview of what entering would show.
	ChildCount int
}

// Options controls how a directory is listed.
type Options struct {
	// ShowHidden includes dotfiles in the listing.
	ShowHidden bool
	// DirsOnly skips regular files.
	DirsOnly bool
	// UseGitignore filters entries matched by the nearest .gitignore (and
	// its ancestors up to the git root).
	UseGitignore bool
}

// List reads dir, applies the requested filters, and returns a slice sorted
// with directories first, then by case-insensitive name. The "." and ".."
// entries are never included; the caller can synthesize them.
func List(dir string, opts Options) ([]Entry, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}

	var matcher *gitignoreMatcher
	if opts.UseGitignore {
		matcher = loadGitignore(abs)
	}

	out := make([]Entry, 0, len(raw))
	for _, e := range raw {
		name := e.Name()
		if !opts.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		path := filepath.Join(abs, name)
		info, err := e.Info()
		if err != nil {
			continue
		}

		if matcher != nil && matcher.match(path, info.IsDir()) {
			continue
		}
		if opts.DirsOnly && !info.IsDir() {
			continue
		}

		entry := Entry{
			Name:       name,
			Path:       path,
			IsDir:      info.IsDir(),
			IsSymlink:  info.Mode()&os.ModeSymlink != 0,
			Mode:       info.Mode(),
			Size:       info.Size(),
			ChildCount: -1,
		}
		if entry.IsSymlink {
			if t, err := os.Readlink(path); err == nil {
				entry.Target = t
			}
		}
		if entry.IsDir {
			if children, err := os.ReadDir(path); err == nil {
				entry.ChildCount = len(children)
			}
		}
		out = append(out, entry)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}
