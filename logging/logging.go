// Package logging configures the global stdlib logger to write to a rolling
// file under the user's config directory while still echoing to stdout. In a
// packaged Windows GUI build stdout is discarded, so the file is the only
// durable record of backend activity; in `wails dev` the terminal still gets
// everything.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"grails/cmdutil"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxSizeMB    = 5
	maxBackups   = 5
	maxAgeDays   = 30
	logFileName  = "grails.log"
	logSubFolder = "logs"
)

// Setup wires the stdlib logger to a rolling file plus stdout. Returns the
// directory where log files live so the UI can offer "open logs folder".
func Setup() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating log dir %s: %w", dir, err)
	}

	rotator := &lumberjack.Logger{
		Filename:   filepath.Join(dir, logFileName),
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   true,
	}

	log.SetOutput(io.MultiWriter(os.Stdout, rotator))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("[logging] writing to %s (rotation: %dMB / %d backups / %d days)",
		rotator.Filename, maxSizeMB, maxBackups, maxAgeDays)
	return dir, nil
}

// Dir returns the directory where log files live (does not create it).
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}
	return filepath.Join(base, "grails", logSubFolder), nil
}

// OpenFolder opens the given directory in the OS file browser. On Windows
// this spawns explorer.exe (which exits with code 1 even on success — we use
// Start() rather than Run() so that's irrelevant).
func OpenFolder(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	cmdutil.HideWindow(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	return nil
}
