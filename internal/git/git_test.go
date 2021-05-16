package git

import (
	"os"
	"strings"
	"testing"

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
	tempdir(t)
	gitInit(t)
	gitCommit(t, "chore: foobar")
	gitCommit(t, "lalalala")
	gitTag(t, "v1.2.3")
	createBranch(t, "not-main")
	gitCommit(t, "docs: update")
	gitCommit(t, "foo: bar")
	gitTag(t, "v1.2.4")
	checkout(t, "-")

	t.Run("normal", func(t *testing.T) {
		is := is.New(t)
		tag, err := DescribeTag("")
		is.NoErr(err)
		is.Equal("v1.2.3", tag)
	})

	t.Run("all-branches", func(t *testing.T) {
		is := is.New(t)
		tag, err := DescribeTag("all-branches")
		is.NoErr(err)
		is.Equal("v1.2.4", tag)
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

func checkout(tb testing.TB, branch string) {
	is := is.New(tb)
	_, err := run("checkout", branch)
	is.NoErr(err)
}

func createBranch(tb testing.TB, branch string) {
	is := is.New(tb)
	_, err := run("checkout", "-b", branch)
	is.NoErr(err)
}

func gitTag(tb testing.TB, tag string) {
	is := is.New(tb)
	_, err := run("tag", tag)
	is.NoErr(err)
}

func gitCommit(tb testing.TB, msg string) {
	is := is.New(tb)
	_, err := run("commit", "--allow-empty", "-am", msg)
	is.NoErr(err)
}

func gitInit(tb testing.TB) {
	is := is.New(tb)
	_, err := run("init")
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
