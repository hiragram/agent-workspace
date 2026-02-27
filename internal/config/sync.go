package config

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// syncFiles is the list of individual files to sync from claudeHome.
var syncFiles = []string{"settings.json", "CLAUDE.md"}

// syncDirs is the list of directories to sync from claudeHome.
var syncDirs = []string{"hooks", "plugins", "commands", "agents"}

// Syncer syncs host Claude settings to the container-side config directory.
type Syncer interface {
	SyncSettings(claudeHome, containerClaudeHome string) error
	EnsureOnboardingState(path string) error
}

// DefaultSyncer is the default implementation of Syncer.
type DefaultSyncer struct{}

// NewSyncer creates a new DefaultSyncer.
func NewSyncer() *DefaultSyncer {
	return &DefaultSyncer{}
}

// SyncSettings copies settings files and directories from claudeHome to containerClaudeHome.
func (s *DefaultSyncer) SyncSettings(claudeHome, containerClaudeHome string) error {
	if err := os.MkdirAll(containerClaudeHome, 0755); err != nil {
		return fmt.Errorf("creating container claude home: %w", err)
	}

	for _, f := range syncFiles {
		src := filepath.Join(claudeHome, f)
		dst := filepath.Join(containerClaudeHome, f)
		if err := copyFileIfExists(src, dst); err != nil {
			return fmt.Errorf("syncing file %s: %w", f, err)
		}
	}

	for _, d := range syncDirs {
		src := filepath.Join(claudeHome, d)
		dst := filepath.Join(containerClaudeHome, d)
		if err := syncDirIfExists(src, dst); err != nil {
			return fmt.Errorf("syncing directory %s: %w", d, err)
		}
	}

	return nil
}

// copyFileIfExists copies src to dst if src exists. Does nothing if src doesn't exist.
func copyFileIfExists(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// syncDirIfExists removes dst and copies src to dst recursively, if src exists.
func syncDirIfExists(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	// Remove old destination
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("removing old %s: %w", dst, err)
	}

	return copyDir(src, dst)
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFileIfExists(path, target)
	})
}
