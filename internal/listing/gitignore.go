package listing

import (
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// gitignoreMatcher resolves paths against the .gitignore set rooted at the
// nearest enclosing .git directory. Patterns from each ancestor .gitignore
// (root → leaf) are concatenated and compiled together; they are matched
// against the path relative to the git root, the way git does.
type gitignoreMatcher struct {
	root    string
	matcher *ignore.GitIgnore
}

// loadGitignore walks upward from start to find a .git directory, then
// gathers every .gitignore between that root and start. Returns nil when
// no git root is found or no .gitignore files exist.
func loadGitignore(start string) *gitignoreMatcher {
	root := findGitRoot(start)
	if root == "" {
		return nil
	}

	files := gitignoreFilesBetween(root, start)
	if len(files) == 0 {
		return nil
	}

	var lines []string
	// Always ignore the .git directory itself.
	lines = append(lines, ".git/")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		lines = append(lines, strings.Split(string(data), "\n")...)
	}

	return &gitignoreMatcher{
		root:    root,
		matcher: ignore.CompileIgnoreLines(lines...),
	}
}

// match reports whether path is ignored. isDir is required because gitignore
// patterns ending in "/" only match directories.
func (m *gitignoreMatcher) match(path string, isDir bool) bool {
	rel, err := filepath.Rel(m.root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}
	if isDir {
		rel += "/"
	}
	return m.matcher.MatchesPath(rel)
}

// findGitRoot walks up from start looking for a .git directory. Returns the
// containing directory or "" if no git repo is found before reaching the FS root.
func findGitRoot(start string) string {
	dir := start
	for {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// gitignoreFilesBetween returns existing .gitignore files from root down to
// (and including) leaf, in root-first order. That order matters: deeper
// patterns can override shallower ones.
func gitignoreFilesBetween(root, leaf string) []string {
	var dirs []string
	dir := leaf
	for {
		dirs = append([]string{dir}, dirs...)
		if dir == root {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	var files []string
	for _, d := range dirs {
		p := filepath.Join(d, ".gitignore")
		if _, err := os.Stat(p); err == nil {
			files = append(files, p)
		}
	}
	return files
}
