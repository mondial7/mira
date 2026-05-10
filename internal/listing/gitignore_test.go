package listing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot_AscendsToFindRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got := findGitRoot(deep)
	if got != root {
		t.Errorf("findGitRoot(%q) = %q, want %q", deep, got, root)
	}
}

func TestFindGitRoot_NoRepoReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	if got := findGitRoot(dir); got != "" {
		t.Errorf("expected empty for non-repo, got %q", got)
	}
}

func TestGitignoreMatcher_RespectsPatterns(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, ".gitignore"),
		[]byte("dist/\n*.log\n!keep.log\n"),
		0o644,
	); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}

	m := loadGitignore(root)
	if m == nil {
		t.Fatal("expected matcher to load")
	}

	cases := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{filepath.Join(root, "dist"), true, true},
		{filepath.Join(root, "src"), true, false},
		{filepath.Join(root, "a.log"), false, true},
		{filepath.Join(root, "keep.log"), false, false},
		{filepath.Join(root, "main.go"), false, false},
	}
	for _, tc := range cases {
		got := m.match(tc.path, tc.isDir)
		if got != tc.ignored {
			t.Errorf("match(%q dir=%v) = %v, want %v", tc.path, tc.isDir, got, tc.ignored)
		}
	}
}

func TestGitignoreMatcher_AlwaysIgnoresDotGit(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Empty .gitignore so the only effective rule is the baked-in .git/.
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	m := loadGitignore(root)
	if m == nil {
		t.Fatal("expected matcher to load")
	}
	if !m.match(filepath.Join(root, ".git"), true) {
		t.Error(".git directory should always be ignored")
	}
}

func TestLoadGitignore_NoRepoReturnsNil(t *testing.T) {
	if got := loadGitignore(t.TempDir()); got != nil {
		t.Errorf("expected nil matcher outside repo, got %+v", got)
	}
}
