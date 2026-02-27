package main

import (
	"os"

	"github.com/hiragram/claude-docker/internal/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
