package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		version string
		want    *SemVer
	}{
		{"v1.2.3", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: ""}},
		{"1.2.3", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: ""}},
		{"1.2.3-alpha.1+build.1", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "alpha.1", build: "build.1"}},
		{"v1.0.0-alpha", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "alpha", build: ""}},
		{"v1.0.0-alpha.1", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "alpha.1", build: ""}},
		{"v1.0.0-alpha.beta", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "alpha.beta", build: ""}},
		{"v1.0.0-beta", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "beta", build: ""}},
		{"v1.0.0-beta.2", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "beta.2", build: ""}},
		{"v1.0.0-beta.11", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "beta.11", build: ""}},
		{"v1.0.0-rc.1", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "rc.1", build: ""}},
		{"v1.0.0", &SemVer{major: 1, minor: 0, patch: 0, preRelease: "", build: ""}},
		{"v1.2.0", &SemVer{major: 1, minor: 2, patch: 0, preRelease: "", build: ""}},
		{"v1.2.3-456", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "456", build: ""}},
		{"v1.2.3-456.789", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "456.789", build: ""}},
		{"v1.2.3-456-789", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "456-789", build: ""}},
		{"v1.2.3-456a", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "456a", build: ""}},
		{"v1.2.3-pre", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "pre", build: ""}},
		{"v1.2.3-pre+meta", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "pre", build: "meta"}},
		{"v1.2.3-pre.1", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "pre.1", build: ""}},
		{"v1.2.3-zzz", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "zzz", build: ""}},
		{"v1.2.3", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: ""}},
		{"v1.2.3+meta", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: "meta"}},
		{"v1.2.3+meta-pre", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: "meta-pre"}},
		{"v1.2.3+meta-pre.sha.256a", &SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: "meta-pre.sha.256a"}},
	}
	for _, tt := range tests {
		got, _ := Parse(tt.version)
		if *got != *tt.want {
			t.Errorf("Invalid output :: want %+v, got :: %+v", tt.want, got)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		version string
		want    any
	}{
		{"hello", "invalid semantic version string"},
		{"v1-alpha.beta.gamma", "invalid semantic version string"},
		{"v1-pre", "invalid semantic version string"},
		{"v1+build", "invalid semantic version string"},
		{"v1-pre+build", "invalid semantic version string"},
		{"v1.2-pre+meta", "invalid semantic version string"},
		{"v1", "invalid semantic version string"},
		{"v1.0", "invalid semantic version string"},
		{"a.b.c", "invalid semantic version string"},
		{"v1.2", "invalid semantic version string"},
	}
	for _, tt := range tests {
		_, err := Parse(tt.version)
		if err.Error() != tt.want {
			t.Errorf("Error :: got %t, expected %t", err, tt.want)
		}
	}

}

func TestCompareRaw(t *testing.T) {
	tests := []struct {
		name string
		ver1 string
		ver2 string
		want any
	}{
		{
			name: "TestCompareSemver_1",
			ver1: "v1.2.3",
			ver2: "v1.2.4",
			want: -1,
		},
		{
			name: "TestCompareSemver_2",
			ver1: "v1.0.0-alpha",
			ver2: "v1.0.0-alpha.1",
			want: -1,
		},
		{
			name: "TestCompareSemver_3",
			ver1: "v1.0.0-alpha.1",
			ver2: "v1.0.0-alpha",
			want: 1,
		},
		{
			name: "TestCompareSemver_4",
			ver1: "v1.2.5",
			ver2: "v1.2.4",
			want: 1,
		},
		{
			name: "TestCompareSemver_5",
			ver1: "v1.3.5",
			ver2: "v1.2.4",
			want: 1,
		},
		{
			name: "TestCompareSemver_6",
			ver1: "v2.3.5",
			ver2: "v1.2.4",
			want: 1,
		},
		{
			name: "TestCompareSemver_7",
			ver1: "v1.2.5",
			ver2: "v1.3.4",
			want: -1,
		},
		{
			name: "TestCompareSemver_8",
			ver1: "v1.2.5",
			ver2: "v2.3.4",
			want: -1,
		},
		{
			name: "TestCompareSemver_9",
			ver1: "1.2.3-alpha",
			ver2: "1.2.3-beta",
			want: -1,
		},
		{
			name: "TestCompareSemver_10",
			ver1: "1.2.3-beta",
			ver2: "1.2.3-alpha",
			want: 1,
		},
		{
			name: "TestCompareSemver_10",
			ver1: "1.2.3-beta",
			ver2: "1.2.3",
			want: -1,
		},
		{
			name: "TestCompareSemver_11",
			ver1: "1.2.3",
			ver2: "1.2.3-pre",
			want: 1,
		},
		{
			name: "TestCompareSemver_12",
			ver1: "a.b.c",
			ver2: "e.f.g",
			want: "invalid semantic version string",
		},
		{
			name: "TestCompareSemver_13",
			ver1: "1.2.5",
			ver2: "e.f.g",
			want: "invalid semantic version string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareRaw(tt.ver1, tt.ver2)
			if err != nil {
				if err.Error() != tt.want {
					t.Errorf("Error :: got %t, expected %t", err, tt.want)
				}
			} else {
				if got != tt.want {
					t.Errorf("Error in comparing version :: got %d, want %d", got, tt.want)
				}
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		ver1 *SemVer
		ver2 *SemVer
		want any
	}{
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: ""},
			&SemVer{major: 1, minor: 2, patch: 4, preRelease: "", build: ""},
			-1,
		},
		{&SemVer{major: 1, minor: 2, patch: 4, preRelease: "", build: ""},
			&SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: ""},
			1,
		},
		{&SemVer{major: 1, minor: 2, patch: 4, preRelease: "", build: ""},
			&SemVer{major: 1, minor: 2, patch: 4, preRelease: "", build: ""},
			0,
		},
	}
	for _, tt := range tests {
		got, _ := tt.ver1.Compare(tt.ver2)
		if got != tt.want {
			t.Errorf("Invalid output :: want %+v, got :: %+v", tt.want, got)
		}
	}
}

func TestNextMajor(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    any
	}{
		{
			name:    "TestGetNextMajor_1",
			version: "v1.2.3",
			want:    "2.0.0",
		},
		{
			name:    "TestGetNextMajor_2",
			version: "v9.1.1",
			want:    "10.0.0",
		},
		{"TestGetNextMajor_3", "v1.0", "invalid semantic version string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, err := Parse(tt.version)
			if err != nil {
				if err.Error() != tt.want {
					t.Errorf("Error :: got %t, expected %t", err, tt.want)
				}
			} else {
				got := ver.NextMajor().String()
				if got != tt.want {
					t.Errorf("invalid next major :: got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNextMinor(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    any
	}{
		{
			name:    "TestGetNextMinor_1",
			version: "v1.2.3",
			want:    "1.3.0",
		},
		{
			name:    "TestGetNextMinor_2",
			version: "v9.1.1",
			want:    "9.2.0",
		},
		{"TestGetNextMinor_3", "v1.0", "invalid semantic version string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, err := Parse(tt.version)
			if err != nil {
				if err.Error() != tt.want {
					t.Errorf("Error :: got %t, expected %t", err, tt.want)
				}
			} else {
				got := ver.NextMinor().String()
				if got != tt.want {
					t.Errorf("invalid next minor :: got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNextPatch(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    any
	}{
		{
			name:    "TestGetNextPatch_1",
			version: "v1.2.3",
			want:    "1.2.4",
		},
		{
			name:    "TestGetNextPatch_2",
			version: "v9.1.1",
			want:    "9.1.2",
		},
		{"TestGetNextPatch_3", "v1.0", "invalid semantic version string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, err := Parse(tt.version)
			if err != nil {
				if err.Error() != tt.want {
					t.Errorf("Error :: got %t, expected %t", err, tt.want)
				}
			} else {
				got := ver.NextPatch().String()
				if got != tt.want {
					t.Errorf("invalid next patch :: got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestIsPreRelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "TestIsPreRelease_1",
			version: "v1.2.3-beta.1",
			want:    true,
		},
		{
			name:    "TestIsPreRelease_2",
			version: "v1.2.3",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, err := Parse(tt.version)
			if err != nil {
				t.Errorf("Error :: got %t, expected %t", err, tt.want)
			}
			got := ver.IsPreRelease()
			if tt.want != got {
				t.Errorf("Error in testing IsPreRelease :: got %t, want %t", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		version *SemVer
		want    string
	}{
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: ""}, "1.2.3"},
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "rc.1", build: ""}, "1.2.3-rc.1"},
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: "SNAPSHOT"}, "1.2.3+SNAPSHOT"},
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "rc.1", build: "SNAPSHOT"}, "1.2.3-rc.1+SNAPSHOT"},
	}
	for _, tt := range tests {
		got := tt.version.String()
		if got != tt.want {
			t.Errorf("Invalid output :: want %+v, got :: %+v", tt.want, got)
		}
	}
}

func TestSemVer_PreRelease(t *testing.T) {
	tests := []struct {
		version *SemVer
		tag     string
		want    string
	}{
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "", build: "SNAPSHOT"}, "pre01", "1.2.3-pre01+SNAPSHOT"},
		{&SemVer{major: 1, minor: 2, patch: 3, preRelease: "rc.1", build: "SNAPSHOT"}, "pre02", "1.2.3-pre02+SNAPSHOT"},
	}
	for _, tt := range tests {
		got := tt.version.PreRelease(tt.tag).String()
		if got != tt.want {
			t.Errorf("Invalid output :: want %+v, got :: %+v", tt.want, got)
		}
	}
}
