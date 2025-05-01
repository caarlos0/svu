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

// Commit is a commit with a hash, title (first line of the message), and body
// (rest of the message, not including the title).
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

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

func Root() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	wt, err := repo.Worktree()
	if err != nil {
		return ""
	}

	return wt.Filesystem.Root()
}

type versionSorter struct {
	tags     []string
	versions []*version.Version
}

func (v *versionSorter) Len() int {
	return len(v.tags)
}

func (v *versionSorter) Swap(i, j int) {
	v.tags[i], v.tags[j] = v.tags[j], v.tags[i]
	v.versions[i], v.versions[j] = v.versions[j], v.versions[i]
}

func (v *versionSorter) Less(i, j int) bool {
	return v.versions[i].LessThan(v.versions[j])
}

func getAllTags(tagMode string) ([]string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	var tags []string
	var versions []*version.Version

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagName := strings.TrimPrefix(tagRef.Name().String(), "refs/tags/")

		if tagMode == TagModeCurrent {
			// Check if tag is reachable from HEAD
			head, err := repo.Head()
			if err != nil {
				return err
			}

			commit, err := repo.CommitObject(head.Hash())
			if err != nil {
				return err
			}

			tagCommit, err := repo.CommitObject(tagRef.Hash())
			if err != nil {
				// Skip if it's not a commit (e.g., annotated tag)
				return nil
			}

			isAncestor, err := commit.IsAncestor(tagCommit)
			if err != nil || !isAncestor {
				return nil
			}
		}

		// Try to parse as version
		v, err := version.NewVersion(tagName)
		if err != nil {
			// If not a version, we'll sort these alphabetically after versions
			v, _ = version.NewVersion("0.0.0")
		}

		tags = append(tags, tagName)
		versions = append(versions, v)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort tags by version (descending)
	sorter := &versionSorter{
		tags:     tags,
		versions: versions,
	}
	sort.Sort(sort.Reverse(sorter))

	return sorter.tags, nil
}

func DescribeTag(tagMode string, pattern string) (string, error) {
	tags, err := getAllTags(tagMode)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}
	if pattern == "" {
		return tags[0], nil
	}

	g, err := glob.Compile(pattern)
	if err != nil {
		return "", err
	}
	for _, tag := range tags {
		if g.Match(tag) {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no tags match '%s'", pattern)
}

func Changelog(tag string, dirs []string) ([]Commit, error) {
	if tag == "" {
		return gitLog(dirs, plumbing.ZeroHash)
	} else {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return nil, err
		}

		tagRef, err := repo.Tag(tag)
		if err != nil {
			return nil, err
		}

		// Resolve the tag to a commit hash
		var tagHash plumbing.Hash
		switch tagRef.Hash().Type() {
		case plumbing.CommitObject:
			tagHash = tagRef.Hash()
		case plumbing.TagObject:
			tagObj, err := repo.TagObject(tagRef.Hash())
			if err != nil {
				return nil, err
			}
			tagHash = tagObj.Target
		default:
			return nil, fmt.Errorf("unsupported tag type: %s", tagRef.Hash().Type())
		}

		return gitLog(dirs, tagHash)
	}
}

func gitLog(dirs []string, since plumbing.Hash) ([]Commit, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	var commitIter object.CommitIter
	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}

	if since == plumbing.ZeroHash {
		commitIter, err = repo.Log(&git.LogOptions{From: headRef.Hash()})
	} else {
		commitIter, err = repo.Log(&git.LogOptions{
			From:  headRef.Hash(),
			Since: &since,
		})
	}
	if err != nil {
		return nil, err
	}
	defer commitIter.Close()

	var result []Commit
	for {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Skip if we've reached the since commit
		if since != plumbing.ZeroHash && commit.Hash == since {
			break
		}

		// Filter by directories if specified
		if len(dirs) > 0 {
			files, err := commit.Files()
			if err != nil {
				return nil, err
			}

			var found bool
			err = files.ForEach(func(file *object.File) error {
				for _, dir := range dirs {
					if strings.HasPrefix(file.Name, dir) {
						found = true
						return io.EOF // break early
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
			commit.Hash.String(),
			title,
			body,
		})
	}

	return result, nil
}
