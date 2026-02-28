package container

import (
	"testing"

	"github.com/hiragram/claude-docker/internal/docker"
)

func TestBuildRunConfig_Basic(t *testing.T) {
	mounts := []docker.Mount{
		{Source: "vol", Target: "/data", IsVolume: true},
	}

	config := BuildRunConfig(RunOptions{
		ImageName:  "claude-code-docker",
		Mounts:     mounts,
		ClaudeHome: "/home/user/.claude",
		WorkDir:    "/home/user/project",
		CLIArgs:    nil,
	})

	if config.ImageName != "claude-code-docker" {
		t.Errorf("ImageName = %q, want %q", config.ImageName, "claude-code-docker")
	}

	if config.WorkDir != "/home/user/project" {
		t.Errorf("WorkDir = %q, want %q", config.WorkDir, "/home/user/project")
	}

	if config.EnvVars["HOST_CLAUDE_HOME"] != "/home/user/.claude" {
		t.Errorf("HOST_CLAUDE_HOME = %q, want %q", config.EnvVars["HOST_CLAUDE_HOME"], "/home/user/.claude")
	}

	if config.EnvVars["HOST_WORKSPACE"] != "/home/user/project" {
		t.Errorf("HOST_WORKSPACE = %q, want %q", config.EnvVars["HOST_WORKSPACE"], "/home/user/project")
	}

	// Command should be ["claude", "--allow-dangerously-skip-permissions"]
	if len(config.Command) != 2 {
		t.Fatalf("Command length = %d, want 2", len(config.Command))
	}
	if config.Command[0] != "claude" {
		t.Errorf("Command[0] = %q, want %q", config.Command[0], "claude")
	}
	if config.Command[1] != "--allow-dangerously-skip-permissions" {
		t.Errorf("Command[1] = %q, want %q", config.Command[1], "--allow-dangerously-skip-permissions")
	}

	if len(config.Mounts) != 1 {
		t.Errorf("Mounts length = %d, want 1", len(config.Mounts))
	}
}

func TestBuildRunConfig_WithCLIArgs(t *testing.T) {
	config := BuildRunConfig(RunOptions{
		ImageName:  "claude-code-docker",
		ClaudeHome: "/home/user/.claude",
		WorkDir:    "/workspace",
		CLIArgs:    []string{"-p", "explain this codebase"},
	})

	// Command should be ["claude", "-p", "explain this codebase", "--allow-dangerously-skip-permissions"]
	expected := []string{"claude", "-p", "explain this codebase", "--allow-dangerously-skip-permissions"}
	if len(config.Command) != len(expected) {
		t.Fatalf("Command length = %d, want %d", len(config.Command), len(expected))
	}
	for i, arg := range expected {
		if config.Command[i] != arg {
			t.Errorf("Command[%d] = %q, want %q", i, config.Command[i], arg)
		}
	}
}

func TestBuildRunConfig_EmptyCLIArgs(t *testing.T) {
	config := BuildRunConfig(RunOptions{
		ImageName:  "claude-code-docker",
		ClaudeHome: "/home/user/.claude",
		WorkDir:    "/workspace",
		CLIArgs:    []string{},
	})

	// Command should be ["claude", "--allow-dangerously-skip-permissions"]
	if len(config.Command) != 2 {
		t.Fatalf("Command length = %d, want 2", len(config.Command))
	}
	if config.Command[0] != "claude" {
		t.Errorf("Command[0] = %q, want %q", config.Command[0], "claude")
	}
}

func TestBuildRunConfig_MountsPassedThrough(t *testing.T) {
	mounts := []docker.Mount{
		{Source: "vol1", Target: "/a", IsVolume: true},
		{Source: "/host/b", Target: "/b"},
		{Source: "/host/c", Target: "/c", ReadOnly: true},
	}

	config := BuildRunConfig(RunOptions{
		ImageName:  "img",
		Mounts:     mounts,
		ClaudeHome: "/home/.claude",
		WorkDir:    "/work",
	})

	if len(config.Mounts) != 3 {
		t.Fatalf("Mounts length = %d, want 3", len(config.Mounts))
	}

	// Verify mounts are passed through exactly
	for i, m := range mounts {
		if config.Mounts[i] != m {
			t.Errorf("Mounts[%d] = %+v, want %+v", i, config.Mounts[i], m)
		}
	}
}
