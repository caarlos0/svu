package svu

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/caarlos0/svu/v2/internal/git"
	"github.com/stretchr/testify/require"
)

func TestIsBreaking(t *testing.T) {
	for _, commit := range []git.Commit{
		{Title: "feat!: foo"},
		{Title: "chore(lala)!: foo"},
		{Title: "docs: lalala", Body: "BREAKING CHANGE: lalal"},
		{Title: "docs: lalala", Body: "BREAKING-CHANGE: lalal"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, isBreaking(commit)) // should be a major change
		})
	}

	for _, commit := range []git.Commit{
		{Title: "feat: foo"},
		{Title: "chore(lol): foo"},
		{Title: "docs: lalala"},
		{Title: "docs: BREAKING change: lalal"},
		{Title: "docs: breaking-change: aehijhk"},
		{Title: "docs: BREAKING_CHANGE: foo"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, !isBreaking(commit)) // should NOT be a major change
		})
	}
}

func TestIsFeature(t *testing.T) {
	for _, commit := range []git.Commit{
		{Title: "feat: foo"},
		{Title: "feat(lalal): foobar"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, isFeature(commit)) // should be a minor change
		})
	}

	for _, commit := range []git.Commit{
		{Title: "fix: foo"},
		{Title: "chore: foo"},
		{Title: "docs: lalala"},
		{Title: "ci: foo"},
		{Title: "test: foo"},
		{Title: "Merge remote-tracking branch 'origin/main'"},
		{Title: "refactor: foo bar"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, !isFeature(commit)) // should NOT be a minor change
		})
	}
}

func TestIsPatch(t *testing.T) {
	for _, commit := range []git.Commit{
		{Title: "fix: foo"},
		{Title: "fix(lalal): lalala"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, isPatch(commit)) // should be a patch change
		})
	}

	for _, commit := range []git.Commit{
		{Title: "chore: foobar"},
		{Title: "docs: something"},
		{Title: "invalid commit"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, !isPatch(commit)) // should NOT be a patch change
		})
	}
}

func TestFindNext(t *testing.T) {
	version0a := semver.MustParse("v0.4.5")
	version0b := semver.MustParse("v0.5.5")
	version1 := semver.MustParse("v1.2.3")
	version2 := semver.MustParse("v2.4.12")
	version3 := semver.MustParse("v3.4.5-beta34+ads")
	for expected, next := range map[string]semver.Version{
		"v0.4.5": findNext(version0a, false, false, []git.Commit{{Title: "chore: should do nothing"}}),
		"v0.4.6": findNext(version0a, false, false, []git.Commit{{Title: "fix: inc patch"}}),
		"v0.5.0": findNext(version0a, false, false, []git.Commit{{Title: "feat: inc minor"}}),
		"v1.0.0": findNext(version0b, false, false, []git.Commit{{Title: "feat!: inc minor"}}),
		"v0.6.0": findNext(version0b, true, false, []git.Commit{{Title: "feat!: inc minor"}}),
		"v1.2.3": findNext(version1, false, false, []git.Commit{{Title: "chore: should do nothing"}}),
		"v1.2.4": findNext(version1, false, true, []git.Commit{{Title: "chore: is forcing patch, so should inc patch"}}),
		"v1.3.0": findNext(version1, false, false, []git.Commit{{Title: "feat: inc major"}}),
		"v2.0.0": findNext(version1, false, true, []git.Commit{{Title: "chore!: hashbang incs major"}}),
		"v3.0.0": findNext(version2, false, false, []git.Commit{{Title: "feat: something", Body: "BREAKING CHANGE: increases major"}}),
		"v3.5.0": findNext(version3, false, false, []git.Commit{{Title: "feat: inc major"}}),
	} {
		t.Run(expected, func(t *testing.T) {
			require.True(t, semver.MustParse(expected).Equal(&next)) // expected and next version should match
		})
	}
}

func TestCmd(t *testing.T) {
	ver := func() *semver.Version { return semver.MustParse("1.2.3-pre+123") }
	t.Run(CurrentCmd, func(t *testing.T) {
		cmd := CurrentCmd
		t.Run("version has meta", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.3-pre+123", v.String())
		})
		t.Run("version is clean", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("v1.2.3"), "v1.2.3", "doesnt matter", "nope", nil, false, true)
			require.NoError(t, err)
			require.Equal(t, "1.2.3", v.String())
		})
	})

	t.Run(MinorCmd, func(t *testing.T) {
		cmd := MinorCmd
		t.Run("no meta", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.3.0", v.String())
		})
		t.Run("build", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "", "124", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.3.0+124", v.String())
		})
		t.Run("prerel", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "alpha.1", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.3.0-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "alpha.2", "125", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.3.0-alpha.2+125", v.String())
		})
	})

	t.Run(PatchCmd, func(t *testing.T) {
		cmd := PatchCmd
		t.Run("no meta", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3"), "v1.2.3", "", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.4", v.String())
		})
		t.Run("previous had meta", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3", "", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.3", v.String())
		})
		t.Run("previous had meta, force", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3", "", "", nil, false, true)
			require.NoError(t, err)
			require.Equal(t, "1.2.4", v.String())
		})
		t.Run("previous had meta, force, add meta", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3-alpha.1+1", "alpha.2", "10", nil, false, true)
			require.NoError(t, err)
			require.Equal(t, "1.2.4-alpha.2+10", v.String())
		})
		t.Run("previous had meta, change it", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3-alpha.1+1", "alpha.2", "10", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.3-alpha.2+10", v.String())
		})
		t.Run("build", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3"), "v1.2.3", "", "124", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.4+124", v.String())
		})
		t.Run("prerel", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3"), "v1.2.3", "alpha.1", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.4-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			v, err := nextVersion(cmd, semver.MustParse("1.2.3"), "v1.2.3", "alpha.2", "125", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "1.2.4-alpha.2+125", v.String())
		})
	})

	t.Run(MajorCmd, func(t *testing.T) {
		cmd := MajorCmd
		t.Run("no meta", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "2.0.0", v.String())
		})
		t.Run("build", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "", "124", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "2.0.0+124", v.String())
		})
		t.Run("prerel", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "alpha.1", "", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "2.0.0-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			v, err := nextVersion(cmd, ver(), "v1.2.3", "alpha.2", "125", nil, false, false)
			require.NoError(t, err)
			require.Equal(t, "2.0.0-alpha.2+125", v.String())
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("invalid build", func(t *testing.T) {
			_, err := nextVersion(MinorCmd, semver.MustParse("1.2.3"), "v1.2.3", "", "+125", nil, false, false)
			require.True(t, err != nil)
		})
		t.Run("invalid prerelease", func(t *testing.T) {
			_, err := nextVersion(MinorCmd, semver.MustParse("1.2.3"), "v1.2.3", "+aaa", "", nil, false, false)
			require.True(t, err != nil)
		})
	})
}

func Test_nextPreRelease(t *testing.T) {
	type args struct {
		current    *semver.Version
		next       *semver.Version
		preRelease string
	}
	tests := []struct {
		name    string
		args    args
		want    semver.Version
		wantErr bool
	}{
		{
			name: "no current suffix and no suffix supplied",
			args: args{
				current:    semver.MustParse("1.2.3"),
				next:       semver.MustParse("1.3.0"),
				preRelease: "",
			},
			want:    *semver.MustParse("1.3.0"),
			wantErr: true,
		},
		{
			name: "supplied suffix overrides current suffix",
			args: args{
				current:    semver.MustParse("1.2.3-alpha.1"),
				next:       semver.MustParse("1.3.0"),
				preRelease: "beta",
			},
			want:    *semver.MustParse("1.3.0-beta.0"),
			wantErr: false,
		},
		{
			name: "current suffix is incremented",
			args: args{
				current:    semver.MustParse("1.2.3-alpha.11"),
				next:       semver.MustParse("1.2.3"),
				preRelease: "",
			},
			want:    *semver.MustParse("1.2.3-alpha.12"),
			wantErr: false,
		},
		{
			name: "current suffix is incremented when supplied suffix matches current",
			args: args{
				current:    semver.MustParse("1.2.3-alpha.11"),
				next:       semver.MustParse("1.2.3"),
				preRelease: "alpha",
			},
			want:    *semver.MustParse("1.2.3-alpha.12"),
			wantErr: false,
		},
		{
			name: "pre release version resets if next version changes",
			args: args{
				current:    semver.MustParse("1.2.3-alpha.11"),
				next:       semver.MustParse("1.2.4"),
				preRelease: "alpha",
			},
			want:    *semver.MustParse("1.2.4-alpha.0"),
			wantErr: false,
		},
		{
			name: "increments a current tag that has build metadata",
			args: args{
				current:    semver.MustParse("1.2.3-alpha.1+build.43"),
				next:       semver.MustParse("1.2.3"),
				preRelease: "",
			},
			want:    *semver.MustParse("1.2.3-alpha.2"),
			wantErr: false,
		},
		{
			name: "don't increment if explicit pre-release is supplied",
			args: args{
				current:    semver.MustParse("1.2.3-alpha.1"),
				next:       semver.MustParse("1.2.3"),
				preRelease: "alpha.10",
			},
			want:    *semver.MustParse("1.2.3-alpha.10"),
			wantErr: false,
		},
		{
			name: "prerelease suffix contains a number",
			args: args{
				current:    semver.MustParse("1.2.3-alpha123.1"),
				next:       semver.MustParse("1.2.3"),
				preRelease: "alpha123",
			},
			want:    *semver.MustParse("1.2.3-alpha123.2"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nextPreRelease(tt.args.current, tt.args.next, tt.args.preRelease)
			if tt.wantErr {
				if err == nil {
					t.Errorf("nextPreRelease() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nextPreRelease() = %v, want %v", got, tt.want)
			}
		})
	}
}
