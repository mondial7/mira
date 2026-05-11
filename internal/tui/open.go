package tui

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openPathFunc launches path with the OS-default handler. It's a package
// variable so tests can swap in a stub without spawning real processes.
var openPathFunc = openPath

// openPath hands path off to the platform's default opener and returns
// once the child has been started — we don't wait for the GUI app to
// exit, since that would freeze the TUI.
func openPath(path string) error {
	var (
		bin  string
		args []string
	)
	switch runtime.GOOS {
	case "darwin":
		bin, args = "open", []string{path}
	case "windows":
		bin, args = "rundll32", []string{"url.dll,FileProtocolHandler", path}
	default:
		bin, args = "xdg-open", []string{path}
	}
	cmd := exec.Command(bin, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: %w", bin, err)
	}
	// Release the child so it isn't reaped against the TUI's process group.
	go func() { _ = cmd.Wait() }()
	return nil
}
