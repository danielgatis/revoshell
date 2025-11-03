package version

import "testing"

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "default dev version",
			version:  "dev",
			expected: "dev",
		},
		{
			name:     "release version",
			version:  "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "empty version",
			version:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := Version

			defer func() { Version = original }()

			// Set test value
			Version = tt.version

			got := GetVersion()
			if got != tt.expected {
				t.Errorf("GetVersion() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetGitCommit(t *testing.T) {
	tests := []struct {
		name      string
		gitCommit string
		expected  string
	}{
		{
			name:      "default unknown commit",
			gitCommit: "unknown",
			expected:  "unknown",
		},
		{
			name:      "short commit hash",
			gitCommit: "abc1234",
			expected:  "abc1234",
		},
		{
			name:      "full commit hash",
			gitCommit: "abc1234567890def1234567890abcdef12345678",
			expected:  "abc1234567890def1234567890abcdef12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := GitCommit

			defer func() { GitCommit = original }()

			// Set test value
			GitCommit = tt.gitCommit

			got := GetGitCommit()
			if got != tt.expected {
				t.Errorf("GetGitCommit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetBuildDate(t *testing.T) {
	tests := []struct {
		name      string
		buildDate string
		expected  string
	}{
		{
			name:      "default unknown date",
			buildDate: "unknown",
			expected:  "unknown",
		},
		{
			name:      "ISO 8601 date",
			buildDate: "2024-11-03T10:00:00Z",
			expected:  "2024-11-03T10:00:00Z",
		},
		{
			name:      "custom format date",
			buildDate: "2024-11-03_10:00:00",
			expected:  "2024-11-03_10:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := BuildDate

			defer func() { BuildDate = original }()

			// Set test value
			BuildDate = tt.buildDate

			got := GetBuildDate()
			if got != tt.expected {
				t.Errorf("GetBuildDate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetFullVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		gitCommit string
		expected  string
	}{
		{
			name:      "dev version includes commit",
			version:   "dev",
			gitCommit: "abc1234",
			expected:  "dev-abc1234",
		},
		{
			name:      "dev version with unknown commit",
			version:   "dev",
			gitCommit: "unknown",
			expected:  "dev-unknown",
		},
		{
			name:      "release version excludes commit",
			version:   "v1.0.0",
			gitCommit: "abc1234",
			expected:  "v1.0.0",
		},
		{
			name:      "release version with tag",
			version:   "v2.5.1",
			gitCommit: "def5678",
			expected:  "v2.5.1",
		},
		{
			name:      "pre-release version",
			version:   "v1.0.0-rc1",
			gitCommit: "abc1234",
			expected:  "v1.0.0-rc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			originalVersion := Version
			originalCommit := GitCommit

			defer func() {
				Version = originalVersion
				GitCommit = originalCommit
			}()

			// Set test values
			Version = tt.version
			GitCommit = tt.gitCommit

			got := GetFullVersion()
			if got != tt.expected {
				t.Errorf("GetFullVersion() = %v, want %v", got, tt.expected)
			}
		})
	}
}
