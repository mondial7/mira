package listing

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// scaffold builds a deterministic temp tree:
//
//	root/
//	  .git/                  (so gitignore lookup finds a root)
//	  .gitignore             (ignores: build/, *.log)
//	  README.md
//	  main.go
//	  build/
//	    out.bin
//	  src/
//	    a.go
//	    b.go
//	  app.log
//	  .secret                (dotfile)
//
// It returns the absolute root path. All file contents are empty.
func scaffold(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	mkdirs := []string{
		".git",
		"build",
		"src",
	}
	for _, d := range mkdirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	files := map[string]string{
		".gitignore": "build/\n*.log\n",
		"README.md":  "",
		"main.go":    "",
		"build/out.bin": "",
		"src/a.go":   "",
		"src/b.go":   "",
		"app.log":    "",
		".secret":    "",
	}
	for name, body := range files {
		full := filepath.Join(root, name)
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

func names(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func TestList_DefaultHidesDotfiles(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range entries {
		if e.Name == ".secret" || e.Name == ".gitignore" || e.Name == ".git" {
			t.Errorf("dotfile %q should be hidden by default", e.Name)
		}
	}
}

func TestList_ShowHiddenIncludesDotfiles(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{ShowHidden: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got := map[string]bool{}
	for _, e := range entries {
		got[e.Name] = true
	}
	for _, want := range []string{".gitignore", ".secret"} {
		if !got[want] {
			t.Errorf("expected %q with ShowHidden, got %v", want, names(entries))
		}
	}
}

func TestList_GitignoreFilters(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{UseGitignore: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range entries {
		if e.Name == "build" {
			t.Errorf("build/ should be gitignored")
		}
		if e.Name == "app.log" {
			t.Errorf("app.log should be gitignored")
		}
	}
}

func TestList_NoGitignoreShowsEverything(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{UseGitignore: false})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got := map[string]bool{}
	for _, e := range entries {
		got[e.Name] = true
	}
	for _, want := range []string{"build", "app.log"} {
		if !got[want] {
			t.Errorf("expected %q without gitignore, got %v", want, names(entries))
		}
	}
}

func TestList_DirsOnly(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{DirsOnly: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir {
			t.Errorf("non-dir %q returned with DirsOnly", e.Name)
		}
	}
}

func TestList_SortsDirsFirstThenAlpha(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{UseGitignore: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	seenFile := false
	for _, e := range entries {
		if !e.IsDir {
			seenFile = true
			continue
		}
		if seenFile {
			t.Errorf("directory %q after a file in sorted output: %v", e.Name, names(entries))
		}
	}

	// Within each group entries should be alphabetical (case-insensitive).
	var prev string
	for _, e := range entries {
		if e.IsDir {
			continue
		}
		lower := strings.ToLower(e.Name)
		if prev != "" && lower < prev {
			t.Errorf("file order broken: %q after %q", e.Name, prev)
		}
		prev = lower
	}
}

func TestList_PopulatesEntryFields(t *testing.T) {
	root := scaffold(t)
	entries, err := List(root, Options{UseGitignore: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Errorf("empty Name in %+v", e)
		}
		if !filepath.IsAbs(e.Path) {
			t.Errorf("Path should be absolute, got %q", e.Path)
		}
	}
}

func TestList_SymlinkDetection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink test requires unix-style symlinks")
	}
	root := scaffold(t)
	target := filepath.Join(root, "main.go")
	link := filepath.Join(root, "main.symlink")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	entries, err := List(root, Options{UseGitignore: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	var found *Entry
	for i := range entries {
		if entries[i].Name == "main.symlink" {
			found = &entries[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("symlink entry missing from listing: %v", names(entries))
	}
	if !found.IsSymlink {
		t.Errorf("entry not flagged as symlink: %+v", *found)
	}
	if found.Target != target {
		t.Errorf("symlink target = %q, want %q", found.Target, target)
	}
}

func TestList_NonexistentReturnsError(t *testing.T) {
	_, err := List("/this/should/not/exist/banana-four", Options{})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
