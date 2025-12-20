package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion_DefaultValue(t *testing.T) {
	// Test that default version is returned when not set via ldflags
	version := GetVersion()
	assert.NotEmpty(t, version)
	assert.Contains(t, version, "dev")
}

func TestGetBuildTime_DefaultValue(t *testing.T) {
	// Test that default build time is returned when not set via ldflags
	buildTime := GetBuildTime()
	assert.NotEmpty(t, buildTime)
	assert.Equal(t, "unknown", buildTime)
}

func TestGetGitCommit_DefaultValue(t *testing.T) {
	// Test that default git commit is returned when not set via ldflags
	commit := GetGitCommit()
	assert.NotEmpty(t, commit)
	assert.Equal(t, "unknown", commit)
}

func TestGetInfo_ReturnsCompleteInfo(t *testing.T) {
	// Test that GetInfo returns all build information
	info := GetInfo()

	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.BuildTime)
	assert.NotEmpty(t, info.GitCommit)
	assert.Contains(t, info.Version, "dev")
	assert.Equal(t, "unknown", info.BuildTime)
	assert.Equal(t, "unknown", info.GitCommit)
}

func TestInfo_String(t *testing.T) {
	// Test that Info.String() returns formatted string
	info := Info{
		Version:   "1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	str := info.String()
	assert.Contains(t, str, "1.0.0")
	assert.Contains(t, str, "2024-01-01T00:00:00Z")
	assert.Contains(t, str, "abc123")
}

func TestSetVersion(t *testing.T) {
	// Test that we can override version programmatically (for testing)
	originalVersion := version
	defer func() { version = originalVersion }()

	testVersion := "2.0.0-test"
	setVersionForTest(testVersion)

	assert.Equal(t, testVersion, GetVersion())
}

func TestSetBuildTime(t *testing.T) {
	// Test that we can override build time programmatically (for testing)
	originalBuildTime := buildTime
	defer func() { buildTime = originalBuildTime }()

	testTime := "2024-01-01T12:00:00Z"
	setBuildTimeForTest(testTime)

	assert.Equal(t, testTime, GetBuildTime())
}

func TestSetGitCommit(t *testing.T) {
	// Test that we can override git commit programmatically (for testing)
	originalCommit := gitCommit
	defer func() { gitCommit = originalCommit }()

	testCommit := "def456"
	setGitCommitForTest(testCommit)

	assert.Equal(t, testCommit, GetGitCommit())
}
