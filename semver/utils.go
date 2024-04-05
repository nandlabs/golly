package semver

import (
	"fmt"
	"strings"
)

// compare returns an integer with 3 possible values, -1, 0, +1
func compare(c1, c2 *SemVer) (int, error) {
	// compare major version
	if c1.major != c2.major {
		if c1.major > c2.major {
			return 1, nil
		} else {
			return -1, nil
		}
	}

	// compare minor version
	if c1.minor != c2.minor {
		if c1.minor > c2.minor {
			return 1, nil
		} else {
			return -1, nil
		}
	}

	// compare patch version
	if c1.patch != c2.patch {
		if c1.patch > c2.patch {
			return 1, nil
		} else {
			return -1, nil
		}
	}
	return comparePreRelease(c1.preRelease, c2.preRelease)
}

func comparePreRelease(v1, v2 string) (int, error) {

	pre1 := len(v1) > 1
	pre2 := len(v2) > 1

	if pre1 && pre2 {
		return strings.Compare(v1, v2), nil
	}
	if pre1 {
		return -1, nil
	}

	if pre2 {
		return 1, nil
	}

	return 0, fmt.Errorf("no pre-release versions present")
}
