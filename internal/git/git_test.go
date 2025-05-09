package git

import (
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testGitRepo struct {
	repo *git.Repository
}

func (t *testGitRepo) PlainOpen(path string) (*git.Repository, error) {
	return t.repo, nil
}

func setupTestRepo(t *testing.T) *testGitRepo {
	t.Helper()
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	sig := &object.Signature{
		Name:  "test",
		Email: "test@example.com",
		When:  time.Now(),
	}

	writeFile := func(fs billy.Filesystem, filename, content string) {
		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write([]byte(content))
		require.NoError(t, err)
	}

	filename := "file.txt"
	writeFile(wt.Filesystem, filename, "Initial content")

	_, err = wt.Add(filename)
	require.NoError(t, err)

	hash, err := wt.Commit("Initial commit\n\nBody message", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	_, err = repo.CreateTag("v1.0.0", hash, nil)
	require.NoError(t, err)

	writeFile(wt.Filesystem, filename, "Second commit content")

	_, err = wt.Add(filename)
	require.NoError(t, err)

	hash2, err := wt.Commit("Second commit\n\nMore details", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	_, err = repo.CreateTag("v1.1.0", hash2, &git.CreateTagOptions{
		Tagger:  sig,
		Message: "annotated tag",
	})
	require.NoError(t, err)

	writeFile(wt.Filesystem, filename, "Third commit content")

	_, err = wt.Add(filename)
	require.NoError(t, err)

	_, err = wt.Commit("Third commit\n\nWith more changes", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	return &testGitRepo{repo: repo}
}

func TestIsRepo(t *testing.T) {
	repo := setupTestRepo(t)
	g := &Git{open: repo.PlainOpen}

	isRepo, err := g.IsRepo()
	require.NoError(t, err)
	assert.True(t, isRepo)
}

func TestRoot(t *testing.T) {
	repo := setupTestRepo(t)
	g := &Git{open: repo.PlainOpen}

	root, err := g.Root()
	require.NoError(t, err)
	assert.NotEmpty(t, root)
}

func TestGetAllTags(t *testing.T) {
	repo := setupTestRepo(t)
	g := &Git{open: repo.PlainOpen}

	tags, err := g.getAllTags(TagModeAll)
	require.NoError(t, err)
	assert.Equal(t, []string{"v1.1.0", "v1.0.0"}, tags)
}

func TestDescribeTag(t *testing.T) {
	repo := setupTestRepo(t)
	g := &Git{open: repo.PlainOpen}

	t.Run("no pattern", func(t *testing.T) {
		tag, err := g.DescribeTag(TagModeAll, "")
		require.NoError(t, err)
		assert.Equal(t, "v1.1.0", tag)
	})

	t.Run("with pattern", func(t *testing.T) {
		tag, err := g.DescribeTag(TagModeAll, "v1.0.*")
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", tag)
	})

	t.Run("no match", func(t *testing.T) {
		tag, err := g.DescribeTag(TagModeAll, "v2.*")
		assert.Error(t, err)
		assert.Empty(t, tag)
	})
}

func TestChangelog(t *testing.T) {
	repo := setupTestRepo(t)
	g := &Git{open: repo.PlainOpen}

	t.Run("full changelog", func(t *testing.T) {
		commits, err := g.Changelog("", nil)
		require.NoError(t, err)
		require.Len(t, commits, 3)
		assert.Equal(t, "Third commit", commits[0].Title)
		assert.Equal(t, "Initial commit", commits[2].Title)
	})

	t.Run("since tag", func(t *testing.T) {
		commits, err := g.Changelog("v1.0.0", nil)
		require.NoError(t, err)
		require.Len(t, commits, 2)
		assert.Equal(t, "Third commit", commits[0].Title)
		assert.Equal(t, "Second commit", commits[1].Title)
	})
}

func TestGitLog(t *testing.T) {
	repo := setupTestRepo(t)
	g := &Git{open: repo.PlainOpen}

	t.Run("retrieve all commits", func(t *testing.T) {
		commits, err := g.gitLog(nil, plumbing.ZeroHash)
		require.NoError(t, err)
		require.Len(t, commits, 3)
		assert.Equal(t, "Third commit", commits[0].Title)
		assert.Equal(t, "Initial commit", commits[2].Title)
	})

	t.Run("stop at specific commit", func(t *testing.T) {
		// Get the hash of the "Initial commit"
		commits, err := g.gitLog(nil, plumbing.ZeroHash)
		require.NoError(t, err)
		require.Len(t, commits, 3)
		initialCommitHash := commits[2].SHA

		// Use the hash of the "Initial commit" as the `since` parameter
		commits, err = g.gitLog(nil, plumbing.NewHash(initialCommitHash))
		require.NoError(t, err)
		require.Len(t, commits, 2)
		assert.Equal(t, "Third commit", commits[0].Title)
		assert.Equal(t, "Second commit", commits[1].Title)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty repository", func(t *testing.T) {
		repo, err := git.Init(memory.NewStorage(), memfs.New())
		require.NoError(t, err)
		g := &Git{open: func(path string) (*git.Repository, error) { return repo, nil }}

		isRepo, err := g.IsRepo()
		require.NoError(t, err)
		assert.True(t, isRepo)

		tags, err := g.getAllTags(TagModeAll)
		require.NoError(t, err)
		assert.Empty(t, tags)

		commits, err := g.gitLog(nil, plumbing.ZeroHash)
		require.NoError(t, err)
		assert.Empty(t, commits)
	})

	t.Run("no tags", func(t *testing.T) {
		repo := setupTestRepo(t)
		g := &Git{open: repo.PlainOpen}

		// Remove all tags
		iter, err := repo.repo.Tags()
		require.NoError(t, err)
		iter.ForEach(func(ref *plumbing.Reference) error {
			return repo.repo.DeleteTag(ref.Name().Short())
		})

		tags, err := g.getAllTags(TagModeAll)
		require.NoError(t, err)
		assert.Empty(t, tags)
	})
}

func TestDescribeTag_NoTagsFound(t *testing.T) {
	mockOpen := func(path string) (*git.Repository, error) {
		// Return a mock repository or simulate an empty repository
		return nil, plumbing.ErrObjectNotFound
	}

	g := &Git{}
	g.SetOpenFunc(mockOpen)

	tag, err := g.DescribeTag("", "")
	if err == nil || err.Error() != "no tags found in the repository" {
		t.Fatalf("expected error 'no tags found in the repository', got: %v", err)
	}

	if tag != "" {
		t.Fatalf("expected no tag, got: %s", tag)
	}
}
