package svu

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/caarlos0/svu/v3/internal/git"
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
			require.False(t, isBreaking(commit)) // should NOT be a major change
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
			require.False(t, isFeature(commit)) // should NOT be a minor change
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
			require.False(t, isPatch(commit)) // should NOT be a patch change
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
		"0.4.5": findNext(version0a, []git.Commit{{Title: "chore: should do nothing"}}, Options{Ctx: t.Context()}),
		"0.4.6": findNext(version0a, []git.Commit{{Title: "fix: inc patch"}}, Options{Ctx: t.Context()}),
		"0.5.0": findNext(version0a, []git.Commit{{Title: "feat: inc minor"}}, Options{Ctx: t.Context()}),
		"1.0.0": findNext(version0b, []git.Commit{{Title: "feat!: inc minor"}}, Options{Ctx: t.Context()}),
		"0.6.0": findNext(version0b, []git.Commit{{Title: "feat!: inc minor"}}, Options{Ctx: t.Context(), KeepV0: true}),
		"1.2.3": findNext(version1, []git.Commit{{Title: "chore: should do nothing"}}, Options{Ctx: t.Context()}),
		"1.2.4": findNext(version1, []git.Commit{{Title: "chore: always"}}, Options{Ctx: t.Context(), Always: true}),
		"1.3.0": findNext(version1, []git.Commit{{Title: "feat: inc major"}}, Options{Ctx: t.Context()}),
		"2.0.0": findNext(version1, []git.Commit{{Title: "chore!: hashbang incs major"}}, Options{Ctx: t.Context()}),
		"3.0.0": findNext(version2, []git.Commit{{Title: "feat: something", Body: "BREAKING CHANGE: increases major"}}, Options{Ctx: t.Context()}),
		"3.5.0": findNext(version3, []git.Commit{{Title: "feat: inc major"}}, Options{Ctx: t.Context()}),
	} {
		t.Run(expected, func(t *testing.T) {
			require.Equal(t, expected, next.String())
		})
	}
}

func TestCmd(t *testing.T) {
	ver := func() *semver.Version { return semver.MustParse("1.2.3-pre+123") }
	t.Run("current", func(t *testing.T) {
		t.Run("version has meta", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Current,
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.3-pre+123", v.String())
		})
		t.Run("version is clean", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("v1.2.3"), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Current,
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.3", v.String())
		})
	})

	t.Run("minor", func(t *testing.T) {
		t.Run("clean", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Minor,
			})
			require.NoError(t, err)
			require.Equal(t, "1.3.0", v.String())
		})
		t.Run("metadata", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:      t.Context(),
				Action:   Minor,
				Metadata: "124",
			})
			require.NoError(t, err)
			require.Equal(t, "1.3.0+124", v.String())
		})
		t.Run("prerelease", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:        t.Context(),
				Action:     Minor,
				PreRelease: "alpha.1",
			})
			require.NoError(t, err)
			require.Equal(t, "1.3.0-alpha.1", v.String())
		})
		t.Run("all", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:        t.Context(),
				Action:     Minor,
				PreRelease: "alpha.2",
				Metadata:   "125",
			})
			require.NoError(t, err)
			require.Equal(t, "1.3.0-alpha.2+125", v.String())
		})
	})

	t.Run("patch", func(t *testing.T) {
		t.Run("clean", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Patch,
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.4", v.String())
		})
		t.Run("previous had meta", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Patch,
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.3", v.String())
		})
		t.Run("previous had meta + always", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Patch,
				Always: true,
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.4", v.String())
		})
		t.Run("previous had meta + always, add meta", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3-alpha.1+1", Options{
				Ctx:        t.Context(),
				Action:     Patch,
				Always:     true,
				PreRelease: "alpha.2",
				Metadata:   "10",
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.4-alpha.2+10", v.String())
		})
		t.Run("previous had meta, change it", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3-alpha.1+1", Options{
				Ctx:        t.Context(),
				Action:     Patch,
				PreRelease: "alpha.2",
				Metadata:   "10",
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.3-alpha.2+10", v.String())
		})
		t.Run("metadata", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{
				Ctx:      t.Context(),
				Action:   Patch,
				Metadata: "124",
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.4+124", v.String())
		})
		t.Run("prerelease", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{
				Ctx:        t.Context(),
				Action:     Patch,
				PreRelease: "alpha.1",
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.4-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			v, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{
				Ctx:        t.Context(),
				Action:     Patch,
				Metadata:   "125",
				PreRelease: "alpha.2",
			})
			require.NoError(t, err)
			require.Equal(t, "1.2.4-alpha.2+125", v.String())
		})
	})

	t.Run("major", func(t *testing.T) {
		t.Run("no meta", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:    t.Context(),
				Action: Major,
			})
			require.NoError(t, err)
			require.Equal(t, "2.0.0", v.String())
		})
		t.Run("metadata", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:      t.Context(),
				Action:   Major,
				Metadata: "124",
			})
			require.NoError(t, err)
			require.Equal(t, "2.0.0+124", v.String())
		})
		t.Run("prerelease", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:        t.Context(),
				Action:     Major,
				PreRelease: "alpha.1",
			})
			require.NoError(t, err)
			require.Equal(t, "2.0.0-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			v, err := nextVersion(ver(), "v1.2.3", Options{
				Ctx:        t.Context(),
				Action:     Major,
				PreRelease: "alpha.2",
				Metadata:   "125",
			})
			require.NoError(t, err)
			require.Equal(t, "2.0.0-alpha.2+125", v.String())
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("invalid build", func(t *testing.T) {
			_, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{Ctx: t.Context()})
			require.Error(t, err)
		})
		t.Run("invalid prerelease", func(t *testing.T) {
			_, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{Ctx: t.Context()})
			require.Error(t, err)
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
		{
			name: "from normal version, bump base to next patch when creating prerelease",
			args: args{
				current:    semver.MustParse("0.8.0"),
				next:       semver.MustParse("0.8.0"),
				preRelease: "dev",
			},
			want:    *semver.MustParse("0.8.1-dev.0"),
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
