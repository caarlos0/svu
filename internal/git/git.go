package git

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobwas/glob"
	"github.com/hashicorp/go-version"
)

// GitInterface defines the methods that Git must implement.
type GitInterface interface {
	DescribeTag(tagMode string, pattern string) (string, error)
	Changelog(tag string, dirs []string) ([]Commit, error)
	IsRepo() (bool, error)
	Root() (string, error)
}

// Git is the implementation of GitInterface.
type Git struct {
	open func(path string) (*git.Repository, error)
}

func New() *Git {
	return &Git{
		open: git.PlainOpen,
	}
}

type Commit struct {
	SHA   string
	Title string
	Body  string
}

func (c Commit) String() string {
	return c.SHA + ": " + c.Title + "\n" + c.Body
}

const (
	TagModeAll     = "all"
	TagModeCurrent = "current"
)

// SetOpenFunc allows overriding the default open function for testing purposes.
func (g *Git) SetOpenFunc(openFunc func(path string) (*git.Repository, error)) {
	g.open = openFunc
}

func (g *Git) IsRepo() (bool, error) {
	_, err := g.open(".")
	if err != nil {
		return false, fmt.Errorf("not a git repository: %w", err)
	}
	return true, nil
}

func (g *Git) Root() (string, error) {
	repo, err := g.open(".")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve repository root: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve worktree: %w", err)
	}

	return wt.Filesystem.Root(), nil
}

type versionSorter struct {
	tags     []string
	versions []*version.Version
}

func (v *versionSorter) Len() int           { return len(v.tags) }
func (v *versionSorter) Swap(i, j int)      { v.tags[i], v.tags[j] = v.tags[j], v.tags[i] }
func (v *versionSorter) Less(i, j int) bool { return v.versions[i].LessThan(v.versions[j]) }

func (g *Git) getAllTags() ([]string, error) {
	repo, err := g.open(".")
	if err != nil {
		return nil, err
	}

	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	var tagList []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagList = append(tagList, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Parse tags into semantic versions
	var versions []*version.Version
	for _, tag := range tagList {
		ver, err := version.NewVersion(tag)
		if err != nil {
			// Skip tags that are not valid semantic versions
			continue
		}
		versions = append(versions, ver)
	}

	// Sort tags using versionSorter
	sort.Sort(sort.Reverse(&versionSorter{
		tags:     tagList,
		versions: versions,
	}))

	return tagList, nil
}

func (g *Git) DescribeTag(pattern string) (string, error) {
	tags, err := g.getAllTags()
	if err != nil {
		return "", fmt.Errorf("no tags found in the repository")
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in the repository")
	}
	if pattern == "" {
		return tags[0], nil
	}

	matcher, err := glob.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile pattern: %w", err)
	}
	for _, tag := range tags {
		if matcher.Match(tag) {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no tags match '%s', available tags: %v", pattern, tags)
}

func (g *Git) Changelog(tag string, dirs []string) ([]Commit, error) {
	if tag == "" {
		return g.gitLog(dirs, plumbing.ZeroHash)
	}

	repo, err := g.open(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	tagRef, err := repo.Tag(tag)
	if err != nil {
		return nil, fmt.Errorf("failed to find tag '%s': %w", tag, err)
	}

	var tagCommit *object.Commit
	switch obj, err := repo.Object(plumbing.AnyObject, tagRef.Hash()); err {
	case nil:
		switch obj := obj.(type) {
		case *object.Commit:
			tagCommit = obj
		case *object.Tag:
			tagCommit, err = repo.CommitObject(obj.Target)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve tag target: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported tag type: %T", obj)
		}
	case plumbing.ErrObjectNotFound:
		return nil, fmt.Errorf("tag not found: %s", tag)
	default:
		return nil, fmt.Errorf("failed to retrieve tag object: %w", err)
	}

	return g.gitLog(dirs, tagCommit.Hash)
}

func (g *Git) gitLog(dirs []string, since plumbing.Hash) ([]Commit, error) {
	repo, err := g.open(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	headRef, err := repo.Head()
	if err == plumbing.ErrReferenceNotFound {
		// Handle empty repository gracefully
		return []Commit{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get repository head: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commit log: %w", err)
	}
	defer commitIter.Close()

	var result []Commit
	for {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating through commits: %w", err)
		}

		// Stop at the specified commit hash
		if since != plumbing.ZeroHash && commit.Hash == since {
			break
		}

		if len(dirs) > 0 {
			files, err := commit.Files()
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve files for commit %s: %w", commit.Hash.String(), err)
			}

			var found bool
			err = files.ForEach(func(file *object.File) error {
				for _, dir := range dirs {
					if strings.HasPrefix(file.Name, dir) {
						found = true
						return io.EOF
					}
				}
				return nil
			})
			if err != nil && err != io.EOF {
				return nil, err
			}
			if !found {
				continue
			}
		}

		message := commit.Message
		titleEndIdx := strings.Index(message, "\n")
		var title, body string
		if titleEndIdx < 0 {
			title = message
			body = ""
		} else {
			title = message[:titleEndIdx]
			body = message[titleEndIdx+1:]
		}

		result = append(result, Commit{
			SHA:   commit.Hash.String(),
			Title: title,
			Body:  strings.TrimSpace(body),
		})
	}

	return result, nil
}
