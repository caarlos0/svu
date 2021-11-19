package svu

import (
	"regexp"

	"github.com/Masterminds/semver"
)

var (
	breaking     = regexp.MustCompile("(?im).*breaking change:.*")
	breakingBang = regexp.MustCompile(`(?im).*(\w+)(\(.*\))?!:.*`)
	feature      = regexp.MustCompile(`(?im).*feat(\(.*\))?:.*`)
	patch        = regexp.MustCompile(`(?im).*fix(\(.*\))?:.*`)
)

func isBreaking(log string) bool {
	return breaking.MatchString(log) || breakingBang.MatchString(log)
}

func isFeature(log string) bool {
	return feature.MatchString(log)
}

func isPatch(log string) bool {
	return patch.MatchString(log)
}

func FindNext(current *semver.Version, forcePatchIncrement bool, log string) semver.Version {
	if isBreaking(log) {
		return current.IncMajor()
	}

	if isFeature(log) {
		return current.IncMinor()
	}

	if forcePatchIncrement || isPatch(log) {
		return current.IncPatch()
	}

	return *current
}
