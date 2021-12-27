package svu

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/matryer/is"
)

func TestIsBreaking(t *testing.T) {
	for _, log := range []string{
		"feat!: foo",
		"chore(lala)!: foo",
		"docs: lalala\nBREAKING CHANGE: lalal",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(isBreaking(log)) // should be a major change
		})
	}

	for _, log := range []string{
		"feat: foo",
		"chore(lol): foo",
		"docs: lalala",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(!isBreaking(log)) // should NOT be a major change
		})
	}
}

func TestIsFeature(t *testing.T) {
	for _, log := range []string{
		"feat: foo",
		"feat(lalal): foobar",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(isFeature(log)) // should be a minor change
		})
	}

	for _, log := range []string{
		"fix: foo",
		"chore: foo",
		"docs: lalala",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(!isFeature(log)) // should NOT be a minor change
		})
	}
}

func TestIsPatch(t *testing.T) {
	for _, log := range []string{
		"fix: foo",
		"fix(lalal): lalala",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(isPatch(log)) // should be a patch change
		})
	}

	for _, log := range []string{
		"chore: foobar",
		"docs: something",
		"invalid commit",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(!isPatch(log)) // should NOT be a patch change
		})
	}
}

func TestFindNext(t *testing.T) {
	version1 := semver.MustParse("v1.2.3")
	version2 := semver.MustParse("v2.4.12")
	version3 := semver.MustParse("v3.4.5-beta34+ads")
	for expected, next := range map[string]semver.Version{
		"v1.2.3": FindNext(version1, false, "chore: should do nothing"),
		"v1.2.4": FindNext(version1, true, "chore: is forcing patch, so should inc patch"),
		"v1.3.0": FindNext(version1, false, "feat: inc major"),
		"v2.0.0": FindNext(version1, true, "chore!: hashbang incs major"),
		"v3.0.0": FindNext(version2, false, "feat: something\nBREAKING CHANGE: increases major"),
		"v3.5.0": FindNext(version3, false, "feat: inc major"),
	} {
		t.Run(expected, func(t *testing.T) {
			is.New(t).True(semver.MustParse(expected).Equal(&next)) // expected and next version should match
		})
	}
}
