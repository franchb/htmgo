package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a semantic version (https://semver.org) as Go module tags spell
// it: a leading "v", then MAJOR.MINOR.PATCH and an optional -prerelease. Build
// metadata is not allowed in module versions, so it is rejected.
type Version struct {
	Major, Minor, Patch int
	Pre                 string
}

func ParseVersion(s string) (Version, error) {
	rest, ok := strings.CutPrefix(s, "v")
	if !ok {
		return Version{}, fmt.Errorf("version %q: missing leading v", s)
	}
	if strings.Contains(rest, "+") {
		return Version{}, fmt.Errorf("version %q: build metadata is not allowed", s)
	}
	core, pre, hasPre := strings.Cut(rest, "-")
	if hasPre && pre == "" {
		return Version{}, fmt.Errorf("version %q: empty prerelease", s)
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("version %q: want vMAJOR.MINOR.PATCH", s)
	}
	var nums [3]int
	for i, p := range parts {
		if p == "" || (len(p) > 1 && p[0] == '0') {
			return Version{}, fmt.Errorf("version %q: bad number %q", s, p)
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("version %q: bad number %q", s, p)
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Pre: pre}, nil
}

func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// Compare orders versions by MAJOR.MINOR.PATCH, then puts a prerelease before
// its release. Prerelease identifiers are compared as whole strings, which is
// enough here: only stable versions are used as release baselines.
func (v Version) Compare(o Version) int {
	for _, d := range []int{v.Major - o.Major, v.Minor - o.Minor, v.Patch - o.Patch} {
		if d != 0 {
			return d
		}
	}
	switch {
	case v.Pre == o.Pre:
		return 0
	case v.Pre == "":
		return 1
	case o.Pre == "":
		return -1
	}
	return strings.Compare(v.Pre, o.Pre)
}

// Bump is the size of a release, ordered so that a larger bump wins.
type Bump int

const (
	BumpNone Bump = iota
	BumpPatch
	BumpMinor
	BumpMajor
)

func (b Bump) String() string {
	return [...]string{"none", "patch", "minor", "major"}[b]
}

func (v Version) Bump(b Bump) Version {
	switch b {
	case BumpMajor:
		return Version{Major: v.Major + 1}
	case BumpMinor:
		return Version{Major: v.Major, Minor: v.Minor + 1}
	case BumpPatch:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	}
	return v
}
