package git

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestNewRepository(t *testing.T) {
	cwd := currentWorkingDirectory(t)
	cgd := filepath.Join(cwd, ".git")

	t.Run("defaults", func(t *testing.T) {
		is := is.New(t)
		r, err := NewRepository("", "")
		is.NoErr(err)
		is.Equal(r.WorkTree, cwd)
		is.Equal(r.GitDirectory, cgd)
	})

	t.Run("only worktree set", func(t *testing.T) {
		wt := "/idk/some/work/tree/location"
		is := is.New(t)
		r, err := NewRepository(wt, "")
		is.NoErr(err)
		is.Equal(r.WorkTree, wt)
		is.Equal(r.GitDirectory, cgd)
	})

	t.Run("only gitdir set", func(t *testing.T) {
		gd := "/idk/some/git/location"
		is := is.New(t)
		r, err := NewRepository("", gd)
		is.NoErr(err)
		is.Equal(r.WorkTree, cwd)
		is.Equal(r.GitDirectory, gd)
	})

	t.Run("worktree and gitdir set", func(t *testing.T) {
		is := is.New(t)
		r, err := NewRepository(cwd, cgd)
		is.NoErr(err)
		is.Equal(r.WorkTree, cwd)
		is.Equal(r.GitDirectory, cgd)
	})
}

func TestRepository_IsRepo(t *testing.T) {
	t.Run("is not a repo", func(t *testing.T) {
		tempdir(t, true)
		r := Repository{}
		is := is.New(t)
		is.Equal(false, r.IsRepo()) // should not be arepo
	})

	t.Run("is a repo", func(t *testing.T) {
		tempdir(t, true)
		gitInit(t)
		r := Repository{}
		is := is.New(t)
		is.True(r.IsRepo()) // should be arepo
	})
}

func TestRepository_DescribeTag(t *testing.T) {
	setup := func(tb testing.TB) {
		tb.Helper()
		tempdir(tb, true)
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
		r := Repository{}
		tag, err := r.DescribeTag("current-branch", "")
		is.NoErr(err)
		is.Equal("v1.2.4", tag)
	})

	t.Run("all-branches", func(t *testing.T) {
		setup(t)
		is := is.New(t)
		r := Repository{}
		tag, err := r.DescribeTag("all-branches", "")
		is.NoErr(err)
		is.Equal("v1.2.5", tag)
	})

	t.Run("pattern", func(t *testing.T) {
		setup(t)
		is := is.New(t)
		r := Repository{}
		tag, err := r.DescribeTag("current-branch", "pattern-*")
		is.NoErr(err)
		is.Equal("pattern-1.2.3", tag)
	})
}

func TestRepository_Changelog(t *testing.T) {
	tempdir(t, true)
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
	r := Repository{}
	log, err := r.Changelog("v1.2.3", "")
	is.NoErr(err)
	for _, msg := range []string{
		"chore: foobar",
		"fix: foo",
		"feat: foobar",
	} {
		is.True(strings.Contains(log, msg)) // log should contain commit
	}
}

func TestRepository_ChangelogWithDirectory(t *testing.T) {
	tempDir := tempdir(t, true)
	localDir := dir(tempDir, "a-folder", t)
	defer func() { os.RemoveAll(localDir) }()
	file := tempfile(t, localDir, "a-file.txt")
	gitInit(t)
	gitCommit(t, "chore: foobar")
	gitCommit(t, "lalalala")
	gitTag(t, "v1.2.3")
	gitCommit(t, "feat: foobar")
	gitAdd(t, file)
	gitCommit(t, "chore: filtered dir")
	is := is.New(t)
	r := Repository{}
	log, err := r.Changelog("v1.2.3", localDir)
	is.NoErr(err)

	is.True(strings.Contains(log, "chore: filtered dir"))
	is.True(!strings.Contains(log, "feat: foobar"))
}

func TestRepository_run(t *testing.T) {
	// current directory: . , ./.git
	rootWT := currentWorkingDirectory(t)
	rootGD := filepath.Join(rootWT, ".git")

	t.Run("current directory", func(t *testing.T) {
		is := is.New(t)
		rootRepository := Repository{
			WorkTree:     rootWT,
			GitDirectory: rootGD,
		}
		_, err := rootRepository.run("init")
		is.NoErr(err)
		is.True(rootRepository.IsRepo())
		actualRootWT, err := rootRepository.run("rev-parse", "--show-toplevel")
		is.NoErr(err)
		actualRootWT = strings.TrimSuffix(actualRootWT, "\n") // git adds a new line to the cli output
		is.Equal(actualRootWT, rootWT)
		actualGitDir, err := rootRepository.run("rev-parse", "--absolute-git-dir")
		is.NoErr(err)
		actualGitDir = strings.TrimSuffix(actualGitDir, "\n") // git adds a new line to the cli output
		is.Equal(actualGitDir, rootGD)
	})

	// subdirectory: ./foo , ./foo/.git
	subWT := dir(currentWorkingDirectory(t), "foo", t)
	subGD := dir(subWT, ".git", t)

	t.Run("subdirectory", func(t *testing.T) {
		is := is.New(t)
		subfolderRepository := Repository{
			WorkTree:     subWT,
			GitDirectory: subGD,
		}
		_, err := subfolderRepository.run("init")
		is.NoErr(err)
		is.True(subfolderRepository.IsRepo())
		actualWT, err := subfolderRepository.run("rev-parse", "--show-toplevel")
		is.NoErr(err)
		actualWT = strings.TrimSuffix(actualWT, "\n") // git adds a new line to the cli output
		is.Equal(actualWT, subWT)
		actualGitDir, err := subfolderRepository.run("rev-parse", "--absolute-git-dir")
		is.NoErr(err)
		actualGitDir = strings.TrimSuffix(actualGitDir, "\n") // git adds a new line to the cli output
		is.Equal(actualGitDir, subGD)
	})

	// external: /temp/sdfjklds , /temp/sdfjklds/.git
	externalWT := tempdir(t, false)
	externalGD := dir(externalWT, ".git", t)
	tmpWTFileName := "somefile.txt"
	tmpWTFilePath := tempfile(t, externalWT, tmpWTFileName)
	expectedWTContent, _ := dataFromFile(tmpWTFilePath)
	tmpGDFileName := "somefile.txt"
	tmpGDFilePath := tempfile(t, externalGD, tmpGDFileName)
	expectedGDContent, _ := dataFromFile(tmpGDFilePath)

	t.Run("external directory", func(t *testing.T) {
		is := is.New(t)
		externalRepository := Repository{
			WorkTree:     externalWT,
			GitDirectory: externalGD,
		}
		_, err := externalRepository.run("init")
		is.NoErr(err)
		is.True(externalRepository.IsRepo())
		actualWT, err := externalRepository.run("rev-parse", "--show-toplevel")
		is.NoErr(err)
		actualWT = strings.TrimSuffix(actualWT, "\n") // git adds a new line to the cli output
		// compare file content instead of having to deal with symlink paths
		actualWTContent, err := dataFromFile(actualWT + string(os.PathSeparator) + tmpWTFileName)
		is.NoErr(err)
		is.Equal(expectedWTContent, actualWTContent)

		actualGD, err := externalRepository.run("rev-parse", "--absolute-git-dir")
		is.NoErr(err)
		actualGD = strings.TrimSuffix(actualGD, "\n") // git adds a new line to the cli output
		actualGDContent, err := dataFromFile(actualGD + string(os.PathSeparator) + tmpGDFileName)
		is.NoErr(err)
		is.Equal(expectedGDContent, actualGDContent)
	})
}

func dataFromFile(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
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

// tempdir create a temporary directory and optionally cd into it
func tempdir(tb testing.TB, cdToTempDir bool) string {
	is := is.New(tb)
	previous, err := os.Getwd()
	is.NoErr(err)
	dir := tb.TempDir()
	tb.TempDir()
	if cdToTempDir {
		tb.Cleanup(func() {
			is.NoErr(os.Chdir(previous))
		})
		is.NoErr(os.Chdir(dir))
		tb.Logf("cd into %s", dir)
	}
	return dir
}

func dir(tempDir, subfolder string, tb testing.TB) string {
	is := is.New(tb)
	createdDir := path.Join(tempDir, subfolder)
	err := os.Mkdir(createdDir, 0755)
	tb.Cleanup(func() { os.RemoveAll(createdDir) })
	is.NoErr(err)
	return createdDir
}

func tempfile(tb testing.TB, dir, filename string) string {
	is := is.New(tb)
	d1 := []byte("hello\ngo\n")
	file := path.Join(dir, filename)
	err := os.WriteFile(file, d1, 0644)
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
	r := Repository{}
	return r.run(allArgs...)
}

func currentWorkingDirectory(tb *testing.T) string {
	is := is.New(tb)
	getwd, err := os.Getwd()
	is.NoErr(err)
	return getwd
}
