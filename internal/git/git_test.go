package git

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsRepo(t *testing.T) {
	t.Run("is not a repo", func(t *testing.T) {
		tempdir(t)
		require.False(t, IsRepo()) // should not be arepo
	})

	t.Run("is a repo", func(t *testing.T) {
		tempdir(t)
		gitInit(t)
		require.True(t, IsRepo()) // should be arepo
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
	t.Run(TagModeCurrent, func(t *testing.T) {
		setup(t)
		tag, err := DescribeTag(TagModeCurrent, "")
		require.NoError(t, err)
		require.Equal(t, "v1.2.4", tag)
	})

	t.Run(TagModeAll, func(t *testing.T) {
		setup(t)
		tag, err := DescribeTag(TagModeAll, "")
		require.NoError(t, err)
		require.Equal(t, "v1.2.5", tag)
	})

	t.Run("pattern", func(t *testing.T) {
		setup(t)
		tag, err := DescribeTag(TagModeCurrent, "pattern-*")
		require.NoError(t, err)
		require.Equal(t, "pattern-1.2.3", tag)
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
	log, err := Changelog("v1.2.3", nil)
	require.NoError(t, err)
	for _, title := range []string{
		"chore: foobar",
		"fix: foo",
		"feat: foobar",
	} {
		requireLogContains(t, log, title)
	}
}

func requireLogContains(tb testing.TB, log []Commit, title string) {
	tb.Helper()
	for _, commit := range log {
		if commit.Title == title {
			return
		}
	}
	tb.Errorf("expected %v to contain a commit with msg %q", log, title)
}

func requireLogNotContains(tb testing.TB, log []Commit, title string) {
	tb.Helper()
	for _, commit := range log {
		if commit.Title == title {
			tb.Errorf("expected %v to not contain a commit with msg %q", log, title)
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
	log, err := Changelog("v1.2.3", []string{localDir})
	require.NoError(t, err)

	requireLogContains(t, log, "chore: filtered dir")
	requireLogNotContains(t, log, "feat: foobar")
}

func switchToBranch(tb testing.TB, branch string) {
	_, err := fakeGitRun("switch", branch)
	require.NoError(tb, err)
}

func createBranch(tb testing.TB, branch string) {
	_, err := fakeGitRun("switch", "-c", branch)
	require.NoError(tb, err)
}

func gitTag(tb testing.TB, tag string) {
	_, err := fakeGitRun("tag", tag)
	require.NoError(tb, err)
}

func gitCommit(tb testing.TB, msg string) {
	_, err := fakeGitRun("commit", "--allow-empty", "-am", msg)
	require.NoError(tb, err)
}

func gitAdd(tb testing.TB, path string) {
	_, err := fakeGitRun("add", path)
	require.NoError(tb, err)
}

func gitInit(tb testing.TB) {
	_, err := fakeGitRun("init")
	require.NoError(tb, err)
}

func tempdir(tb testing.TB) string {
	previous, err := os.Getwd()
	require.NoError(tb, err)
	tb.Cleanup(func() {
		require.NoError(tb, os.Chdir(previous))
	})
	dir := tb.TempDir()
	require.NoError(tb, os.Chdir(dir))
	tb.Logf("cd into %s", dir)
	return dir
}

func dir(tempDir string, tb testing.TB) string {
	createdDir := path.Join(tempDir, "a-folder")
	err := os.Mkdir(createdDir, 0o755)
	require.NoError(tb, err)
	return createdDir
}

func tempfile(tb testing.TB, dir string) string {
	d1 := []byte("hello\ngo\n")
	file := path.Join(dir, "a-file.txt")
	err := os.WriteFile(file, d1, 0o644)
	require.NoError(tb, err)
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
