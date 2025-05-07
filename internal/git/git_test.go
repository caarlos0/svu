package git

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
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
	repo, err := git.Init(memory.NewStorage(), nil)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	sig := &object.Signature{
		Name:  "test",
		Email: "test@example.com",
		When:  time.Now(),
	}

	// Initial commit
	hash, err := wt.Commit("Initial commit\n\nBody message", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Create tag v1.0.0
	_, err = repo.CreateTag("v1.0.0", hash, nil)
	require.NoError(t, err)

	// Second commit
	hash2, err := wt.Commit("Second commit\n\nMore details", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Create annotated tag v1.1.0
	_, err = repo.CreateTag("v1.1.0", hash2, &git.CreateTagOptions{
		Tagger:  sig,
		Message: "annotated tag",
	})
	require.NoError(t, err)

	// Third commit
	_, err = wt.Commit("Third commit\n\nWith more changes", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	return &testGitRepo{repo: repo}
}

func TestIsRepo(t *testing.T) {
	t.Run("with repository", func(t *testing.T) {
		testRepo := setupTestRepo(t)
		g := &Git{open: testRepo.PlainOpen}
		assert.True(t, g.IsRepo())
	})

	t.Run("without repository", func(t *testing.T) {
		g := &Git{
			open: func(path string) (*git.Repository, error) {
				return nil, fmt.Errorf("not a git repo")
			},
		}
		assert.False(t, g.IsRepo())
	})
}

func TestGetAllTags(t *testing.T) {
	testRepo := setupTestRepo(t)
	g := &Git{open: testRepo.PlainOpen}

	t.Run("all tags", func(t *testing.T) {
		tags, err := g.getAllTags(TagModeAll)
		require.NoError(t, err)
		assert.Len(t, tags, 2)
		assert.Equal(t, "v1.1.0", tags[0])
		assert.Equal(t, "v1.0.0", tags[1])
	})

	t.Run("current tags", func(t *testing.T) {
		tags, err := g.getAllTags(TagModeCurrent)
		require.NoError(t, err)
		assert.Len(t, tags, 2)
	})
}

func TestDescribeTag(t *testing.T) {
	testRepo := setupTestRepo(t)
	g := &Git{open: testRepo.PlainOpen}

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
		_, err := g.DescribeTag(TagModeAll, "v2.*")
		assert.Error(t, err)
	})
}

func TestChangelog(t *testing.T) {
	testRepo := setupTestRepo(t)
	g := &Git{open: testRepo.PlainOpen}

	t.Run("full changelog", func(t *testing.T) {
		commits, err := g.Changelog("", nil)
		require.NoError(t, err)
		assert.Len(t, commits, 3)
		assert.Equal(t, "Third commit", commits[0].Title)
		assert.Equal(t, "Initial commit", commits[2].Title)
	})

	t.Run("since tag", func(t *testing.T) {
		commits, err := g.Changelog("v1.0.0", nil)
		require.NoError(t, err)
		assert.Len(t, commits, 2)
		assert.Equal(t, "Third commit", commits[0].Title)
		assert.Equal(t, "Second commit", commits[1].Title)
	})

	t.Run("invalid tag", func(t *testing.T) {
		_, err := g.Changelog("invalid-tag", nil)
		assert.Error(t, err)
	})
}

func TestRoot(t *testing.T) {
	testRepo := setupTestRepo(t)
	g := &Git{open: testRepo.PlainOpen}

	_, err := g.Root()
	assert.NoError(t, err)
}
