package listing

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

// walkSize sums the byte size of every file inside root recursively. To
// keep listing responsive on huge trees we bail out early once we've
// either visited too many entries or spent too long walking; in that
// case the returned exact flag is false so the caller can label the
// number as approximate.
func walkSize(root string) (bytes int64, exact bool) {
	const (
		walkBudget = 2000
		walkTime   = 30 * time.Millisecond
	)
	deadline := time.Now().Add(walkTime)
	visited := 0
	exact = true

	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip individual unreadable paths instead of failing the whole walk.
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		visited++
		if visited > walkBudget || time.Now().After(deadline) {
			exact = false
			return filepath.SkipAll
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		bytes += info.Size()
		return nil
	})
	return bytes, exact
}

// HumanSize formats a byte count as a short, human-readable string using
// binary (1024-based) units: B, KB, MB, GB, TB, PB, EB.
func HumanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}
