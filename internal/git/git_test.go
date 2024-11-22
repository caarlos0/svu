package git

import (
	"os"
	"path"
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
		gitTag(tb, "v1.2.5-prerelease")
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
	log, err := Changelog("v1.2.3", "")
	is.NoErr(err)
	for _, msg := range []string{
		"chore: foobar",
		"fix: foo",
		"feat: foobar",
	} {
		requireLogContains(t, log, msg)
	}
}

func requireLogContains(tb testing.TB, log []string, msg string) {
	tb.Helper()
	for _, commit := range log {
		if strings.HasSuffix(commit, msg) {
			return
		}
	}
	tb.Errorf("expected %v to contain a commit with msg %q", log, msg)
}

func requireLogNotContains(tb testing.TB, log []string, msg string) {
	tb.Helper()
	for _, commit := range log {
		if strings.HasSuffix(commit, msg) {
			tb.Errorf("expected %v to not contain a commit with msg %q", log, msg)
		}
	}
}

func TestChangelogWithDirectory(t *testing.T) {
	tempDir := tempdir(t)
	localDir := dir(tempDir, t)
	file := tempfile(t, localDir)
	gitInit(t)
	gitCommit(t, "chore: foobar")
	gitCommit(t, "lalalala")
	gitTag(t, "v1.2.3")
	gitCommit(t, "feat: foobar")
	gitAdd(t, file)
	gitCommit(t, "chore: filtered dir")
	is := is.New(t)
	log, err := Changelog("v1.2.3", localDir)
	is.NoErr(err)

	requireLogContains(t, log, "chore: filtered dir")
	requireLogNotContains(t, log, "feat: foobar")
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

func gitAdd(tb testing.TB, path string) {
	is := is.New(tb)
	_, err := fakeGitRun("add", path)
	is.NoErr(err)
}

func gitInit(tb testing.TB) {
	is := is.New(tb)
	_, err := fakeGitRun("init")
	is.NoErr(err)
}

func tempdir(tb testing.TB) string {
	is := is.New(tb)
	previous, err := os.Getwd()
	is.NoErr(err)
	tb.Cleanup(func() {
		is.NoErr(os.Chdir(previous))
	})
	dir := tb.TempDir()
	is.NoErr(os.Chdir(dir))
	tb.Logf("cd into %s", dir)
	return dir
}

func dir(tempDir string, tb testing.TB) string {
	is := is.New(tb)
	createdDir := path.Join(tempDir, "a-folder")
	err := os.Mkdir(createdDir, 0o755)
	is.NoErr(err)
	return createdDir
}

func tempfile(tb testing.TB, dir string) string {
	is := is.New(tb)
	d1 := []byte("hello\ngo\n")
	file := path.Join(dir, "a-file.txt")
	err := os.WriteFile(file, d1, 0o644)
	is.NoErr(err)
	return file
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
