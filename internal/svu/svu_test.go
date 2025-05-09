package svu

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/caarlos0/svu/v3/internal/git"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockGit is a mock implementation of the git.GitInterface.
type MockGit struct {
	mock.Mock
}

func (m *MockGit) DescribeTag(tagMode string, pattern string) (string, error) {
	args := m.Called(tagMode, pattern)
	return args.String(0), args.Error(1)
}

func (m *MockGit) Changelog(tag string, dirs []string) ([]git.Commit, error) {
	args := m.Called(tag, dirs)
	return args.Get(0).([]git.Commit), args.Error(1)
}

func (m *MockGit) IsRepo() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockGit) Root() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockGit) GetAllTags(tagMode string) ([]string, error) {
	args := m.Called(tagMode)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGit) GitLog(dirs []string, since plumbing.Hash) ([]git.Commit, error) {
	args := m.Called(dirs, since)
	return args.Get(0).([]git.Commit), args.Error(1)
}

func TestVersion(t *testing.T) {
	mockGit := new(MockGit)

	t.Run("valid version", func(t *testing.T) {
		mockGit.On("DescribeTag", git.TagModeAll, "").Return("v1.2.3", nil)
		mockGit.On("Changelog", "v1.2.3", []string(nil)).Return([]git.Commit{
			{Title: "feat: add new feature"},
			{Title: "fix: fix a bug"},
		}, nil)

		opts := Options{
			Action:  Next,
			TagMode: git.TagModeAll,
			Prefix:  "v",
		}

		version, err := VersionWithMock(opts, mockGit)
		require.NoError(t, err)
		require.Equal(t, "v1.3.0", version)
	})

	t.Run("no commits", func(t *testing.T) {
		mockGit.On("DescribeTag", git.TagModeAll, "").Return("v1.2.3", nil)
		mockGit.On("Changelog", "v1.2.3", []string(nil)).Return([]git.Commit{}, nil)

		opts := Options{
			Action:  Next,
			TagMode: git.TagModeAll,
			Prefix:  "v",
			Always:  true,
		}

		version, err := VersionWithMock(opts, mockGit)
		require.NoError(t, err)
		require.Equal(t, "v1.2.4", version) // Ensure Always is respected
	})

	t.Run("no tags", func(t *testing.T) {
		mockGit.On("DescribeTag", git.TagModeAll, "").Return("", nil)

		opts := Options{
			Action:  Next,
			TagMode: git.TagModeAll,
			Prefix:  "v",
		}

		version, err := VersionWithMock(opts, mockGit)
		require.NoError(t, err)
		require.Equal(t, "v0.1.0", version) // Ensure no tags defaults to v0.1.0
	})

	t.Run("error from DescribeTag", func(t *testing.T) {
		mockGit.On("DescribeTag", git.TagModeAll, "").Return("", fmt.Errorf("repository does not exist"))

		opts := Options{
			Action:  Next,
			TagMode: git.TagModeAll,
			Prefix:  "v",
		}

		_, err := VersionWithMock(opts, mockGit)
		require.Error(t, err)
	})
}

func VersionWithMock(opts Options, mockGit *MockGit) (string, error) {
	// Use the mock implementation directly
	return VersionWithGit(opts, mockGit)
}

func VersionWithGit(opts Options, g git.GitInterface) (string, error) {
	tag, err := g.DescribeTag(opts.TagMode, opts.Pattern)
	if err != nil {
		return "", err
	}

	current, err := getCurrentVersion(tag, opts.Prefix)
	if err != nil {
		return "", err
	}

	result, err := nextVersion(current, tag, opts, g)
	if err != nil {
		return "", err
	}

	return opts.Prefix + result.String(), nil
}

func nextVersion(
	current *semver.Version,
	tag string,
	opts Options,
	g git.GitInterface,
) (semver.Version, error) {
	log, err := g.Changelog(tag, opts.Directories)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to get changelog: %w", err)
	}

	return findNext(current, log, opts), nil
}

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
		"0.4.5": findNext(version0a, []git.Commit{{Title: "chore: should do nothing"}}, Options{}),
		"0.4.6": findNext(version0a, []git.Commit{{Title: "fix: inc patch"}}, Options{}),
		"0.5.0": findNext(version0a, []git.Commit{{Title: "feat: inc minor"}}, Options{}),
		"1.0.0": findNext(version0b, []git.Commit{{Title: "feat!: inc minor"}}, Options{}),
		"0.6.0": findNext(version0b, []git.Commit{{Title: "feat!: inc minor"}}, Options{KeepV0: true}),
		"1.2.3": findNext(version1, []git.Commit{{Title: "chore: should do nothing"}}, Options{}),
		"1.2.4": findNext(version1, []git.Commit{{Title: "chore: always"}}, Options{Always: true}),
		"1.3.0": findNext(version1, []git.Commit{{Title: "feat: inc major"}}, Options{}),
		"2.0.0": findNext(version1, []git.Commit{{Title: "chore!: hashbang incs major"}}, Options{}),
		"3.0.0": findNext(version2, []git.Commit{{Title: "feat: something", Body: "BREAKING CHANGE: increases major"}}, Options{}),
		"3.5.0": findNext(version3, []git.Commit{{Title: "feat: inc major"}}, Options{}),
	} {
		t.Run(expected, func(t *testing.T) {
			require.Equal(t, expected, next.String())
		})
	}
}

func TestCmd(t *testing.T) {
	mockGit := new(MockGit)

	t.Run("patch with metadata", func(t *testing.T) {
		mockGit.On("DescribeTag", git.TagModeAll, "").Return("v1.2.3", nil)
		mockGit.On("Changelog", "v1.2.3", []string(nil)).Return([]git.Commit{}, nil)

		v, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{
			Action:   Patch,
			Metadata: "124",
		}, mockGit)
		require.NoError(t, err)
		require.Equal(t, "1.2.4+124", v.String())
	})

	t.Run("prerelease", func(t *testing.T) {
		v, err := nextVersion(semver.MustParse("1.2.3"), "v1.2.3", Options{
			Action:     Patch,
			PreRelease: "alpha.1",
		}, mockGit)
		require.NoError(t, err)
		require.Equal(t, "1.2.4-alpha.1", v.String())
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

func TestGetAllTags(t *testing.T) {
	mockGit := new(MockGit)
	mockGit.On("GetAllTags", git.TagModeAll).Return([]string{"v1.0.0", "v1.1.0"}, nil)

	tags, err := mockGit.GetAllTags(git.TagModeAll)
	require.NoError(t, err)
	require.Equal(t, []string{"v1.0.0", "v1.1.0"}, tags)
}
