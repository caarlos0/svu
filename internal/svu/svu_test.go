package svu

import (
	"github.com/caarlos0/svu/v2/internal/git"
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/matryer/is"
)

func TestIsBreaking(t *testing.T) {
	for _, log := range []string{
		"feat!: foo",
		"chore(lala)!: foo",
		"docs: lalala\nBREAKING CHANGE: lalal",
		"docs: lalala\nBREAKING-CHANGE: lalal",
	} {
		t.Run(log, func(t *testing.T) {
			is.New(t).True(isBreaking(log)) // should be a major change
		})
	}

	for _, log := range []string{
		"feat: foo",
		"chore(lol): foo",
		"docs: lalala",
		"docs: BREAKING change: lalal",
		"docs: breaking-change: aehijhk",
		"docs: BREAKING_CHANGE: foo",
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
	version0a := semver.MustParse("v0.4.5")
	version0b := semver.MustParse("v0.5.5")
	version1 := semver.MustParse("v1.2.3")
	version2 := semver.MustParse("v2.4.12")
	version3 := semver.MustParse("v3.4.5-beta34+ads")
	for expected, next := range map[string]semver.Version{
		"v0.4.5": findNext(version0a, false, false, "chore: should do nothing"),
		"v0.4.6": findNext(version0a, false, false, "fix: inc patch"),
		"v0.5.0": findNext(version0a, false, false, "feat: inc minor"),
		"v1.0.0": findNext(version0b, false, false, "feat!: inc minor"),
		"v0.6.0": findNext(version0b, true, false, "feat!: inc minor"),
		"v1.2.3": findNext(version1, false, false, "chore: should do nothing"),
		"v1.2.4": findNext(version1, false, true, "chore: is forcing patch, so should inc patch"),
		"v1.3.0": findNext(version1, false, false, "feat: inc major"),
		"v2.0.0": findNext(version1, false, true, "chore!: hashbang incs major"),
		"v3.0.0": findNext(version2, false, false, "feat: something\nBREAKING CHANGE: increases major"),
		"v3.5.0": findNext(version3, false, false, "feat: inc major"),
	} {
		t.Run(expected, func(t *testing.T) {
			is.New(t).True(semver.MustParse(expected).Equal(&next)) // expected and next version should match
		})
	}
}

func TestCmd(t *testing.T) {
	ver := func() *semver.Version { return semver.MustParse("1.2.3-pre+123") }
	r := &git.Repository{}
	t.Run(CurrentCmd, func(t *testing.T) {
		cmd := CurrentCmd
		t.Run("version has meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "", "", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.3-pre+123", v.String())
		})
		t.Run("version is clean", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("v1.2.3"), "v1.2.3", "doesnt matter", "nope", "", false, true)
			is.NoErr(err)
			is.Equal("1.2.3", v.String())
		})
	})

	t.Run(MinorCmd, func(t *testing.T) {
		cmd := MinorCmd
		t.Run("no meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "", "", "", false, false)
			is.NoErr(err)
			is.Equal("1.3.0", v.String())
		})
		t.Run("build", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "", "124", "", false, false)
			is.NoErr(err)
			is.Equal("1.3.0+124", v.String())
		})
		t.Run("prerel", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "alpha.1", "", "", false, false)
			is.NoErr(err)
			is.Equal("1.3.0-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "alpha.2", "125", "", false, false)
			is.NoErr(err)
			is.Equal("1.3.0-alpha.2+125", v.String())
		})
	})

	t.Run(PatchCmd, func(t *testing.T) {
		cmd := PatchCmd
		t.Run("no meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3"), "v1.2.3", "", "", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.4", v.String())
		})
		t.Run("previous had meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3", "", "", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.3", v.String())
		})
		t.Run("previous had meta, force", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3", "", "", "", false, true)
			is.NoErr(err)
			is.Equal("1.2.4", v.String())
		})
		t.Run("previous had meta, force, add meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3-alpha.1+1", "alpha.2", "10", "", false, true)
			is.NoErr(err)
			is.Equal("1.2.4-alpha.2+10", v.String())
		})
		t.Run("previous had meta, change it", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3-alpha.1+1"), "v1.2.3-alpha.1+1", "alpha.2", "10", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.3-alpha.2+10", v.String())
		})
		t.Run("build", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3"), "v1.2.3", "", "124", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.4+124", v.String())
		})
		t.Run("prerel", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3"), "v1.2.3", "alpha.1", "", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.4-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, semver.MustParse("1.2.3"), "v1.2.3", "alpha.2", "125", "", false, false)
			is.NoErr(err)
			is.Equal("1.2.4-alpha.2+125", v.String())
		})
	})

	t.Run(MajorCmd, func(t *testing.T) {
		cmd := MajorCmd
		t.Run("no meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "", "", "", false, false)
			is.NoErr(err)
			is.Equal("2.0.0", v.String())
		})
		t.Run("build", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "", "124", "", false, false)
			is.NoErr(err)
			is.Equal("2.0.0+124", v.String())
		})
		t.Run("prerel", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "alpha.1", "", "", false, false)
			is.NoErr(err)
			is.Equal("2.0.0-alpha.1", v.String())
		})
		t.Run("all meta", func(t *testing.T) {
			is := is.New(t)
			v, err := nextVersion(r, cmd, ver(), "v1.2.3", "alpha.2", "125", "", false, false)
			is.NoErr(err)
			is.Equal("2.0.0-alpha.2+125", v.String())
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("invalid build", func(t *testing.T) {
			is := is.New(t)
			_, err := nextVersion(r, MinorCmd, semver.MustParse("1.2.3"), "v1.2.3", "", "+125", "", false, false)
			is.True(err != nil)
		})
		t.Run("invalid prerelease", func(t *testing.T) {
			is := is.New(t)
			_, err := nextVersion(r, MinorCmd, semver.MustParse("1.2.3"), "v1.2.3", "+aaa", "", "", false, false)
			is.True(err != nil)
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
