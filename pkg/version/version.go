package version

var (
	// Version is the current version, injected at build time.
	Version = "dev"

	// GitCommit is the git commit hash, injected at build time.
	GitCommit = "unknown"

	// BuildDate is the build date, injected at build time.
	BuildDate = "unknown"
)

// GetVersion returns the current version.
func GetVersion() string {
	return Version
}

// GetGitCommit returns the git commit hash.
func GetGitCommit() string {
	return GitCommit
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return BuildDate
}

// GetFullVersion returns version with commit and date.
func GetFullVersion() string {
	if Version == "dev" {
		return Version + "-" + GitCommit
	}

	return Version
}
