package mount

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectWorktree checks if the .git entry in workDir is a file (indicating a
// git worktree) and returns the main repository's .git directory path.
// Returns empty string if workDir is not a worktree.
func DetectWorktree(workDir string) (string, error) {
	gitPath := filepath.Join(workDir, ".git")

	info, err := os.Lstat(gitPath)
	if err != nil {
		// .git doesn't exist or can't be accessed
		return "", nil
	}

	if info.IsDir() {
		// Regular repository, not a worktree
		return "", nil
	}

	// .git is a file — read the gitdir reference
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return "", err
	}

	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir: ") {
		// Not a valid worktree .git file
		return "", nil
	}

	gitdir := strings.TrimPrefix(line, "gitdir: ")

	// Resolve relative path
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(workDir, gitdir)
	}

	// gitdir points to .git/worktrees/<name>
	// Go up 2 levels to get the main .git directory
	mainGitDir := filepath.Clean(filepath.Join(gitdir, "..", ".."))

	// Resolve to absolute path
	mainGitDir, err = filepath.Abs(mainGitDir)
	if err != nil {
		return "", err
	}

	return mainGitDir, nil
}

// IsSubpath returns true if child is under parent directory.
func IsSubpath(parent, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	if parent == child {
		return true
	}

	return strings.HasPrefix(child, parent+string(filepath.Separator))
}
