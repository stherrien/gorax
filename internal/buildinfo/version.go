package buildinfo

import (
	"fmt"
)

// Build information variables set via ldflags during build
var (
	version   = "dev"     // Version of the application
	buildTime = "unknown" // Build timestamp
	gitCommit = "unknown" // Git commit hash
)

// Info contains build information
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}

// GetVersion returns the application version
func GetVersion() string {
	return version
}

// GetBuildTime returns the build timestamp
func GetBuildTime() string {
	return buildTime
}

// GetGitCommit returns the git commit hash
func GetGitCommit() string {
	return gitCommit
}

// GetInfo returns all build information
func GetInfo() Info {
	return Info{
		Version:   version,
		BuildTime: buildTime,
		GitCommit: gitCommit,
	}
}

// String returns a formatted string of build information
func (i Info) String() string {
	return fmt.Sprintf("Version: %s, Build Time: %s, Git Commit: %s", i.Version, i.BuildTime, i.GitCommit)
}

// Test helpers to set values programmatically
func setVersionForTest(v string) {
	version = v
}

func setBuildTimeForTest(bt string) {
	buildTime = bt
}

func setGitCommitForTest(gc string) {
	gitCommit = gc
}
