package pkg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VersionRange specifies an acceptable range of version for a dependency.
// It has exactly the form "MAJOR.MINOR.PATCH <= v < MAJOR.MINOR.PATCH".
type VersionRange struct {
	Min Version
	Max Version
}

var versionRangeRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+) <= v < (\d+\.\d+\.\d+)$`)

func (vr *VersionRange) UnmarshalText(src []byte) error {
	if !versionRangeRegex.Match(src) {
		return fmt.Errorf("pkg: %q is not a valid version range", string(src))
	}

	versions := versionRangeRegex.FindStringSubmatch(string(src))
	// no need to check the errors if it passes the regex
	vr.Min.UnmarshalText([]byte(versions[1]))
	vr.Max.UnmarshalText([]byte(versions[2]))

	return nil
}

func (vr VersionRange) MarshalText() ([]byte, error) {
	return []byte(vr.String()), nil
}

func (vr VersionRange) String() string {
	return fmt.Sprintf("%s <= v < %s", vr.Min, vr.Max)
}

// Version is a representation of a dependency version of the form
// MAJOR.MINOR.PATCH.
type Version [3]int

var versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

func (v *Version) UnmarshalText(src []byte) error {
	if !versionRegex.Match(src) {
		return fmt.Errorf("pkg: %q is not a valid version", string(src))
	}

	parts := strings.Split(string(src), ".")
	for i, p := range parts {
		// if it passed the regex we can ignore the error
		(*v)[i], _ = strconv.Atoi(p)
	}

	return nil
}

func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}
