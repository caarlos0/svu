package git

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestIsRepo(t *testing.T) {
	t.Run("is not a repo", func(t *testing.T) {
		tempdir(t)
		is := is.New(t)
		is.Equal(false, IsRepo()) // should not be arepo
	})

	t.Run("is a repo", func(t *testing.T) {
		tempdir(t)
		gitInit(t)
		is := is.New(t)
		is.True(IsRepo()) // should be arepo
	})
}

func TestDescribeTag(t *testing.T) {
	setup := func(tb testing.TB) {
		tb.Helper()
		tempdir(tb)
		gitInit(tb)
		gitCommit(tb, "chore: foobar")
		gitTag(tb, "pattern-1.2.3")
		gitCommit(tb, "lalalala")
		gitTag(tb, "v1.2.3")
		gitTag(tb, "v1.2.4") // multiple tags in a single commit
		gitCommit(tb, "chore: aaafoobar")
		gitCommit(tb, "docs: asdsad")
		gitCommit(tb, "fix: fooaaa")
		time.Sleep(time.Second) // TODO: no idea why, but without the sleep sometimes commits are in wrong order
		createBranch(tb, "not-main")
		gitCommit(tb, "docs: update")
		gitCommit(tb, "foo: bar")
		gitTag(tb, "v1.2.5")
		switchToBranch(tb, "-")
	}
	t.Run("normal", func(t *testing.T) {
		setup(t)
		is := is.New(t)
		tag, err := DescribeTag("current-branch", "")
		is.NoErr(err)
		is.Equal("v1.2.4", tag)
	})

	t.Run("all-branches", func(t *testing.T) {
		setup(t)
		is := is.New(t)
		tag, err := DescribeTag("all-branches", "")
		is.NoErr(err)
		is.Equal("v1.2.5", tag)
	})

	t.Run("pattern", func(t *testing.T) {
		setup(t)
		is := is.New(t)
		tag, err := DescribeTag("current-branch", "pattern-*")
		is.NoErr(err)
		is.Equal("pattern-1.2.3", tag)
	})
}

func TestChangelog(t *testing.T) {
	tempdir(t)
	gitInit(t)
	gitCommit(t, "chore: foobar")
	gitCommit(t, "lalalala")
	gitTag(t, "v1.2.3")
	for _, msg := range []string{
		"chore: foobar",
		"fix: foo",
		"feat: foobar",
	} {
		gitCommit(t, msg)
	}
	is := is.New(t)
	log, err := Changelog("v1.2.3")
	is.NoErr(err)
	for _, msg := range []string{
		"chore: foobar",
		"fix: foo",
		"feat: foobar",
	} {
		is.True(strings.Contains(log, msg)) // log should contain commit
	}
}

func switchToBranch(tb testing.TB, branch string) {
	is := is.New(tb)
	_, err := fakeGitRun("switch", branch)
	is.NoErr(err)
}

func createBranch(tb testing.TB, branch string) {
	is := is.New(tb)
	_, err := fakeGitRun("switch", "-c", branch)
	is.NoErr(err)
}

func gitTag(tb testing.TB, tag string) {
	is := is.New(tb)
	_, err := fakeGitRun("tag", tag)
	is.NoErr(err)
}

func gitCommit(tb testing.TB, msg string) {
	is := is.New(tb)
	_, err := fakeGitRun("commit", "--allow-empty", "-am", msg)
	is.NoErr(err)
}

func gitInit(tb testing.TB) {
	is := is.New(tb)
	_, err := fakeGitRun("init")
	is.NoErr(err)
}

func tempdir(tb testing.TB) {
	is := is.New(tb)
	previous, err := os.Getwd()
	is.NoErr(err)
	tb.Cleanup(func() {
		is.NoErr(os.Chdir(previous))
	})
	dir := tb.TempDir()
	is.NoErr(os.Chdir(dir))
	tb.Logf("cd into %s", dir)
}

func fakeGitRun(args ...string) (string, error) {
	allArgs := []string{
		"-c", "user.name='svu'",
		"-c", "user.email='svu@example.com'",
		"-c", "commit.gpgSign=false",
		"-c", "tag.gpgSign=false",
		"-c", "log.showSignature=false",
	}
	allArgs = append(allArgs, args...)
	return run(allArgs...)
}
