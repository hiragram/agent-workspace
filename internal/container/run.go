package container

import (
	"github.com/hiragram/claude-docker/internal/docker"
)

// RunOptions contains the parameters for building a container run configuration.
type RunOptions struct {
	ImageName  string
	Mounts     []docker.Mount
	ClaudeHome string // host-side CLAUDE_HOME (for HOST_CLAUDE_HOME env var)
	WorkDir    string // host working directory
	CLIArgs    []string
}

// BuildRunConfig assembles a docker.RunConfig from the given options.
func BuildRunConfig(opts RunOptions) docker.RunConfig {
	command := []string{"claude"}
	command = append(command, opts.CLIArgs...)
	command = append(command, "--allow-dangerously-skip-permissions")

	return docker.RunConfig{
		ImageName: opts.ImageName,
		Mounts:    opts.Mounts,
		EnvVars: map[string]string{
			"HOST_CLAUDE_HOME": opts.ClaudeHome,
			"HOST_WORKSPACE":   opts.WorkDir,
		},
		WorkDir: opts.WorkDir,
		Command: command,
	}
}
