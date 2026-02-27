package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureOnboardingState_CreatesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-docker.json")

	syncer := NewSyncer()
	if err := syncer.EnsureOnboardingState(path); err != nil {
		t.Fatalf("EnsureOnboardingState() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != "{}\n" {
		t.Errorf("content = %q, want %q", string(content), "{}\n")
	}
}

func TestEnsureOnboardingState_CreatesWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-docker.json")

	// Create empty file
	os.WriteFile(path, []byte(""), 0644)

	syncer := NewSyncer()
	if err := syncer.EnsureOnboardingState(path); err != nil {
		t.Fatalf("EnsureOnboardingState() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != "{}\n" {
		t.Errorf("content = %q, want %q", string(content), "{}\n")
	}
}

func TestEnsureOnboardingState_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-docker.json")

	existing := `{"hasCompletedOnboarding":true}`
	os.WriteFile(path, []byte(existing), 0644)

	syncer := NewSyncer()
	if err := syncer.EnsureOnboardingState(path); err != nil {
		t.Fatalf("EnsureOnboardingState() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != existing {
		t.Errorf("content = %q, want %q (should be preserved)", string(content), existing)
	}
}
