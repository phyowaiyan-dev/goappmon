package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type SemanticVersion struct {
	Major int
	Minor int
	Patch int
}

func ParseSemanticVersion(value string) (SemanticVersion, error) {
	if !versionPattern.MatchString(value) {
		return SemanticVersion{}, errors.New("version must use format major.minor.patch")
	}

	parts := make([]int, 0, 3)
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == '.' {
			part, err := strconv.Atoi(value[start:i])
			if err != nil {
				return SemanticVersion{}, fmt.Errorf("invalid version number: %w", err)
			}
			parts = append(parts, part)
			start = i + 1
		}
	}

	return SemanticVersion{
		Major: parts[0],
		Minor: parts[1],
		Patch: parts[2],
	}, nil
}

func CompareSemanticVersion(a, b string) (int, error) {
	left, err := ParseSemanticVersion(a)
	if err != nil {
		return 0, err
	}
	right, err := ParseSemanticVersion(b)
	if err != nil {
		return 0, err
	}

	switch {
	case left.Major != right.Major:
		if left.Major < right.Major {
			return -1, nil
		}
		return 1, nil
	case left.Minor != right.Minor:
		if left.Minor < right.Minor {
			return -1, nil
		}
		return 1, nil
	case left.Patch != right.Patch:
		if left.Patch < right.Patch {
			return -1, nil
		}
		return 1, nil
	default:
		return 0, nil
	}
}
