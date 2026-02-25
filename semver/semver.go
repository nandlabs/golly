package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	RegexSemver     = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?(\+([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?$`
	RegexPreRelease = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)-([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)`
)

// SemVer is a struct that represents a semantic version
// with major, minor, patch, pre-release and build metadata.
// For more information on semantic versioning, see https://semver.org/spec/v2.0.0.html
type SemVer struct {
	major      int
	minor      int
	patch      int
	preRelease string
	build      string
}

// Major returns the major version of the SemVer struct.
func (s *SemVer) CurrentMajor() int {
	return s.major
}

// Minor returns the minor version of the SemVer struct.
func (s *SemVer) CurrentMinor() int {
	return s.minor
}

// Patch returns the patch version of the SemVer struct.
func (s *SemVer) CurrentPatch() int {
	return s.patch
}

// PreRelease returns the pre-release metadata of the SemVer struct.
func (s *SemVer) CurrentPreRelease() string {
	return s.preRelease
}

// Build returns the build metadata of the SemVer struct.
func (s *SemVer) CurrentBuild() string {
	return s.build
}

func (s *SemVer) IsCurrentPreRelease() bool {
	input := s.String()
	input = strings.TrimPrefix(input, "v")
	input = strings.TrimPrefix(input, " ")
	semverRegex := regexp.MustCompile(RegexPreRelease)
	match := semverRegex.FindStringSubmatch(input)
	return match != nil
}

// New creates a new SemVer struct with the given major, minor, and patch versions.
func (s *SemVer) String() string {
	switch {
	case s.preRelease != "" && s.build != "":
		return fmt.Sprintf("%d.%d.%d-%s+%s", s.major, s.minor, s.patch, s.preRelease, s.build)
	case s.preRelease != "":
		return fmt.Sprintf("%d.%d.%d-%s", s.major, s.minor, s.patch, s.preRelease)
	case s.build != "":
		return fmt.Sprintf("%d.%d.%d+%s", s.major, s.minor, s.patch, s.build)
	}
	return fmt.Sprintf("%d.%d.%d", s.major, s.minor, s.patch)
}

// Parse will parse the given input string and provide a semver struct
// If the input string is not a valid semver string, an error will be returned
func Parse(input string) (*SemVer, error) {
	parsed, err := parse(input)
	return parsed, err
}

// CompareRaw returns three values -1, 0, +1
// -1 denotes ver1 < ver2
// 0 denotes invalid input
// +1 denotes ver1 > ver2
func CompareRaw(ver1, ver2 string) (int, error) {
	c1, err := Parse(ver1)
	if err != nil {
		return 0, err
	}
	c2, err := Parse(ver2)
	if err != nil {
		return 0, err
	}
	ok, err := compare(c1, c2)
	return ok, err
}

// Compare returns three values -1, 0, +1
// -1 denotes ver1 < ver2
// 0 denotes invalid input
// +1 denotes ver1 > ver2
func (s *SemVer) Compare(v *SemVer) (int, error) {
	ok, err := compare(s, v)
	return ok, err
}

// Next Major Increments the major version
func (s *SemVer) NextMajor() *SemVer {
	s.major++
	s.minor = 0
	s.patch = 0
	return s
}

// NextMinor Increments the minor version
func (s *SemVer) NextMinor() *SemVer {

	s.minor++
	s.patch = 0
	return s
}

// NextPatch Increments the patch version
func (s *SemVer) NextPatch() *SemVer {

	s.patch++
	return s
}

// NextPreRelease
func (s *SemVer) NextPreRelease(tag string) *SemVer {
	s.preRelease = tag
	return s
}

func parse(version string) (*SemVer, error) {

	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, " ")

	semverRegex := regexp.MustCompile(RegexSemver)
	match := semverRegex.FindStringSubmatch(version)
	if match == nil {
		return &SemVer{}, fmt.Errorf("invalid semantic version string")
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return &SemVer{}, err
	}

	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return &SemVer{}, err
	}

	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return &SemVer{}, err
	}
	preRelease := ""
	build := ""
	if len(match) > 3 {

		preRelease = match[5]
	}
	if len(match) > 6 {
		build = match[8]

	}

	return &SemVer{
		major:      major,
		minor:      minor,
		patch:      patch,
		preRelease: preRelease,
		build:      build,
	}, nil
}
