package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Mount represents a Docker mount (bind mount or named volume).
type Mount struct {
	Source   string
	Target   string
	ReadOnly bool
	IsVolume bool // true = named volume, false = bind mount
}

// RunConfig holds the configuration for running a Docker container.
type RunConfig struct {
	ImageName string
	Mounts    []Mount
	EnvVars   map[string]string
	WorkDir   string
	Command   []string
}

// Client is the interface for Docker operations.
type Client interface {
	CheckAvailable() error
	Build(ctx context.Context, imageName, contextDir string) error
	VolumeCreate(ctx context.Context, volumeName string) error
	Run(ctx context.Context, config RunConfig) error
}

// ShellClient implements Client by shelling out to the docker CLI.
type ShellClient struct {
	// DockerPath is the path to the docker binary. Defaults to "docker".
	DockerPath string
}

// NewShellClient creates a new ShellClient with default settings.
func NewShellClient() *ShellClient {
	return &ShellClient{DockerPath: "docker"}
}

func (c *ShellClient) dockerCmd() string {
	if c.DockerPath != "" {
		return c.DockerPath
	}
	return "docker"
}

// CheckAvailable verifies that docker is installed and the daemon is running.
func (c *ShellClient) CheckAvailable() error {
	if _, err := exec.LookPath(c.dockerCmd()); err != nil {
		return fmt.Errorf("docker is not installed or not in PATH")
	}

	cmd := exec.Command(c.dockerCmd(), "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon is not running")
	}
	return nil
}

// Build builds a Docker image from the given build context directory.
func (c *ShellClient) Build(ctx context.Context, imageName, contextDir string) error {
	cmd := exec.CommandContext(ctx, c.dockerCmd(), "build", "-t", imageName, contextDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// VolumeCreate creates a named Docker volume (idempotent).
func (c *ShellClient) VolumeCreate(ctx context.Context, volumeName string) error {
	cmd := exec.CommandContext(ctx, c.dockerCmd(), "volume", "create", volumeName)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// BuildRunArgs constructs the docker CLI arguments for a RunConfig.
// This is exported for testing.
func BuildRunArgs(config RunConfig) []string {
	args := []string{"run", "-it", "--rm"}

	for key, val := range config.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	for _, m := range config.Mounts {
		mountArg := fmt.Sprintf("%s:%s", m.Source, m.Target)
		if m.ReadOnly {
			mountArg += ":ro"
		}
		args = append(args, "-v", mountArg)
	}

	if config.WorkDir != "" {
		args = append(args, "--workdir", config.WorkDir)
	}

	args = append(args, config.ImageName)
	args = append(args, config.Command...)

	return args
}

// Run runs a Docker container interactively with the given RunConfig.
func (c *ShellClient) Run(ctx context.Context, config RunConfig) error {
	args := BuildRunArgs(config)
	cmd := exec.CommandContext(ctx, c.dockerCmd(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
