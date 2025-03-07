package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// detectInstallPath detects the install path of the application.
// If the application is installed in /usr/bin or /usr/local/bin,
// the config file is stored in the home directory.
// Otherwise, the config file is stored in the current directory.
func detectInstallPath() string {
	raybotDir := "./.raybot/"
	if path, err := os.Executable(); err == nil {
		dir := filepath.Dir(path)
		if dir == "/usr/bin" || dir == "/usr/local/bin" {
			home, err := os.UserHomeDir()
			if err != nil {
				return raybotDir
			}
			return filepath.Join(home, raybotDir)
		}
	}

	return raybotDir
}

// initDirs initializes the directories of the application.
// Those directories are:
// - log
// - data
func initDirs(rootPath string) error {
	if err := os.MkdirAll(filepath.Join(rootPath, "log"), 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(rootPath, "data"), 0755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	return nil
}
