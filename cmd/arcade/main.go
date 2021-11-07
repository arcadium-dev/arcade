package main

import (
	"os"
	"path/filepath"
	"runtime"

	"arcadium.dev/arcade/internal/arcade"
)

// Build information.
var (
	version string
	branch  string
	commit  string
	date    string
)

func main() {
	os.Exit(
		arcade.New(filepath.Base(os.Args[0]), version, branch, commit, date, runtime.Version()).Start(os.Args),
	)
}
