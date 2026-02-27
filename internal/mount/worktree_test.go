package mount

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectWorktree_RegularRepo(t *testing.T) {
	workDir := t.TempDir()

	// Create .git as a directory (regular repo)
	os.MkdirAll(filepath.Join(workDir, ".git"), 0755)

	result, err := DetectWorktree(workDir)
	if err != nil {
		t.Fatalf("DetectWorktree() error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for regular repo, got %q", result)
	}
}

func TestDetectWorktree_NoGit(t *testing.T) {
	workDir := t.TempDir()

	// No .git at all
	result, err := DetectWorktree(workDir)
	if err != nil {
		t.Fatalf("DetectWorktree() error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string when no .git, got %q", result)
	}
}

func TestDetectWorktree_WorktreeGitFile(t *testing.T) {
	// Simulate a worktree structure:
	// /tmp/main-repo/.git/worktrees/my-worktree/  (main .git dir content)
	// /tmp/worktree/.git  (file pointing to main)

	mainRepo := t.TempDir()
	mainGitDir := filepath.Join(mainRepo, ".git")
	os.MkdirAll(filepath.Join(mainGitDir, "worktrees", "my-worktree"), 0755)

	worktreeDir := t.TempDir()
	gitdirPath := filepath.Join(mainGitDir, "worktrees", "my-worktree")

	// Write .git file with gitdir reference
	os.WriteFile(
		filepath.Join(worktreeDir, ".git"),
		[]byte("gitdir: "+gitdirPath+"\n"),
		0644,
	)

	result, err := DetectWorktree(worktreeDir)
	if err != nil {
		t.Fatalf("DetectWorktree() error: %v", err)
	}

	expected, _ := filepath.Abs(mainGitDir)
	if result != expected {
		t.Errorf("DetectWorktree() = %q, want %q", result, expected)
	}
}

func TestDetectWorktree_RelativeGitdir(t *testing.T) {
	// Create structure where .git file uses a relative path
	baseDir := t.TempDir()

	mainGitDir := filepath.Join(baseDir, "main-repo", ".git")
	os.MkdirAll(filepath.Join(mainGitDir, "worktrees", "wt"), 0755)

	worktreeDir := filepath.Join(baseDir, "worktree")
	os.MkdirAll(worktreeDir, 0755)

	// Relative path from worktree to main .git/worktrees/wt
	relPath, _ := filepath.Rel(worktreeDir, filepath.Join(mainGitDir, "worktrees", "wt"))
	os.WriteFile(
		filepath.Join(worktreeDir, ".git"),
		[]byte("gitdir: "+relPath+"\n"),
		0644,
	)

	result, err := DetectWorktree(worktreeDir)
	if err != nil {
		t.Fatalf("DetectWorktree() error: %v", err)
	}

	expected, _ := filepath.Abs(mainGitDir)
	if result != expected {
		t.Errorf("DetectWorktree() = %q, want %q", result, expected)
	}
}

func TestDetectWorktree_InvalidGitFile(t *testing.T) {
	workDir := t.TempDir()

	// Write .git file without gitdir: prefix
	os.WriteFile(filepath.Join(workDir, ".git"), []byte("not a valid gitdir reference\n"), 0644)

	result, err := DetectWorktree(workDir)
	if err != nil {
		t.Fatalf("DetectWorktree() error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for invalid .git file, got %q", result)
	}
}

func TestIsSubpath(t *testing.T) {
	tests := []struct {
		parent string
		child  string
		want   bool
	}{
		{"/a/b", "/a/b/c", true},
		{"/a/b", "/a/b", true},
		{"/a/b", "/a/bc", false},
		{"/a/b", "/a", false},
		{"/a/b", "/c/d", false},
		{"/a/b/", "/a/b/c", true},
	}

	for _, tt := range tests {
		t.Run(tt.parent+"_"+tt.child, func(t *testing.T) {
			got := IsSubpath(tt.parent, tt.child)
			if got != tt.want {
				t.Errorf("IsSubpath(%q, %q) = %v, want %v", tt.parent, tt.child, got, tt.want)
			}
		})
	}
}
