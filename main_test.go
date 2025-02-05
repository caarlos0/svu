package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildVersion(t *testing.T) {
	t.Run("dev", func(t *testing.T) {
		require.Equal(t, "svu version dev", buildVersion("dev", "", "", ""))
	})

	t.Run("goreleaser built", func(t *testing.T) {
		require.Equal(t, `svu version v1.2.3
commit: a123cd
built at: 2021-01-02
built by: goreleaser`, buildVersion("v1.2.3", "a123cd", "2021-01-02", "goreleaser"))
	})
}
