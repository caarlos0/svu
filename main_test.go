package main

import (
	"reflect"
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

func TestStripPrefixReturnsVersionOnly(t *testing.T) {
	is.New(t).True(getVersion("v2.3.4", "v", "4.5.6", "", true) == "4.5.6")
}

func TestStripPrefixWhenNoPrefixReturnsVersionOnly(t *testing.T) {
	is.New(t).True(getVersion("2.3.4", "v", "4.5.6", "", true) == "4.5.6")
}

func TestNoStripPrefixReturnsPrefixAndVersion(t *testing.T) {
	is.New(t).True(getVersion("v2.3.4", "v", "4.5.6", "", false) == "v4.5.6")
}

func TestSuffix(t *testing.T) {
	is.New(t).True(getVersion("v2.3.4", "v", "4.5.6", "dev", false) == "v4.5.6-dev")
}

func Test_findNextWithSelectCommits(t *testing.T) {
	type args struct {
		current        *semver.Version
		tag            string
		commitPrefixes []string
	}
	tests := []struct {
		name string
		args args
		want semver.Version
	}{
		{
			name: "next-version-with-valid-filter",
			args: args{
				current:        semver.MustParse("v1.7.0"),
				tag:            "v1.7.0",
				commitPrefixes: []string{"fix"},
			},
			want: *semver.MustParse("v1.7.1"),
		},
		{
			name: "next-version-with-valid-filter",
			args: args{
				current:        semver.MustParse("v1.7.0"),
				tag:            "v1.7.0",
				commitPrefixes: []string{"fix", "feat"},
			},
			want: *semver.MustParse("v1.8.0"),
		},
		{
			name: "next-version-with-no-filter",
			args: args{
				current:        semver.MustParse("v1.7.0"),
				tag:            "v1.7.0",
				commitPrefixes: []string{},
			},
			want: *semver.MustParse("v1.8.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findNext(tt.args.current, tt.args.tag, tt.args.commitPrefixes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findNextWithSelectCommits() = %v, want %v", got, tt.want)
			}
		})
	}
}
