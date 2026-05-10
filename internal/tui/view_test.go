package tui

import (
	"strings"
	"testing"

	"github.com/marcomondini/banana-four/internal/listing"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"", 5, ""},
		{"hi", 5, "hi"},
		{"hello world", 5, "hell…"},
		{"abc", 1, "…"},
		{"abc", 0, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "ab…"},
	}
	for _, tc := range cases {
		if got := truncate(tc.in, tc.max); got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestCenterInCell_PadsToWidth(t *testing.T) {
	got := centerInCell("ab", 6)
	if len(got) < 6 {
		t.Errorf("expected width 6, got %q (len %d)", got, len(got))
	}
	if !strings.Contains(got, "ab") {
		t.Errorf("expected content preserved, got %q", got)
	}
}

func TestFlatList_FormatsByEntryType(t *testing.T) {
	entries := []listing.Entry{
		{Name: "src", IsDir: true},
		{Name: "go.mod"},
		{Name: "link", IsSymlink: true, Target: "go.mod"},
	}
	got := FlatList(entries)
	want := "src/\ngo.mod\nlink -> go.mod\n"
	if got != want {
		t.Errorf("FlatList:\n got: %q\nwant: %q", got, want)
	}
}
