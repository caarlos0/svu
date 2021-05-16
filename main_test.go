package main

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/matryer/is"
)

func TestBuildVersion(t *testing.T) {
	t.Run("dev", func(t *testing.T) {
		is.New(t).Equal("svu version dev", buildVersion("dev", "", "", ""))
	})

	t.Run("goreleaser built", func(t *testing.T) {
		is.New(t).Equal(`svu version v1.2.3
commit: a123cd
built at: 2021-01-02
built by: goreleaser`, buildVersion("v1.2.3", "a123cd", "2021-01-02", "goreleaser"))
	})
}

func TestUnsetMetadata(t *testing.T) {
	is.New(t).True(semver.MustParse("v2.3.4").Equal(unsetMetadata(semver.MustParse("v2.3.4-beta+asd123"))))
}
