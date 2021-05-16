package main

import (
	"testing"

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
			is.New(t).True(isPatch(log, false)) // should be a patch change
		})
	}

	for _, log := range []string{
		"chore: foobar",
		"docs: something",
		"invalid commit",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(!isPatch(log, false)) // should NOT be a patch change
		})
	}

	for _, log := range []string{
		"chore: foobar",
		"docs: something",
		"invalid commit",
	} {
		t.Run(log+" (force)", func(t *testing.T) {
			is.New(t).True(isPatch(log, true)) // should NOT be a patch change
		})
	}
}
