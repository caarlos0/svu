package svu

import (
	"regexp"
	"strings"

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

func FilterCommits(log string, prefixes []string) string {
	lines := strings.Split(log, "\n")
	filtered := make([]string, 0)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		commitMsgPrefix := strings.Split(line, " ")
		if len(commitMsgPrefix) < 2 {
			continue
		}

		for _, prefix := range prefixes {
			if strings.HasPrefix(commitMsgPrefix[1], prefix) {
				filtered = append(filtered, line)
			}
		}
	}
	if len(filtered) > 0 {
		return strings.Join(filtered, "\n")
	} else {
		return ""
	}
}
