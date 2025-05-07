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

type Repository interface {
	PlainOpen(path string) (*git.Repository, error)
}

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

func (g *Git) IsRepo() bool {
	_, err := g.open(".")
	return err == nil
}

func (g *Git) Root() (string, error) {
	repo, err := g.open(".")
	if err != nil {
		return "", err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", err
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

func (g *Git) getAllTags(tagMode string) ([]string, error) {
	repo, err := g.open(".")
	if err != nil {
		return nil, err
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	var tagList []*versionSorter

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagName := strings.TrimPrefix(tagRef.Name().String(), "refs/tags/")

		if tagMode == TagModeCurrent {
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
				return nil
			}

			isAncestor, err := commit.IsAncestor(tagCommit)
			if err != nil || !isAncestor {
				return nil
			}
		}

		ver, err := version.NewVersion(tagName)
		if err != nil {
			ver, _ = version.NewVersion("0.0.0")
		}

		tagList = append(tagList, &versionSorter{
			tags:     []string{tagName},
			versions: []*version.Version{ver},
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(&versionSorter{
		tags: func() []string {
			s := make([]string, len(tagList))
			for i, v := range tagList {
				s[i] = v.tags[0]
			}
			return s
		}(),
		versions: func() []*version.Version {
			s := make([]*version.Version, len(tagList))
			for i, v := range tagList {
				s[i] = v.versions[0]
			}
			return s
		}(),
	}))

	result := make([]string, len(tagList))
	for i, ti := range tagList {
		result[i] = ti.tags[0]
	}

	return result, nil
}

func (g *Git) DescribeTag(tagMode string, pattern string) (string, error) {
	tags, err := g.getAllTags(tagMode)
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

func (g *Git) Changelog(tag string, dirs []string) ([]Commit, error) {
	if tag == "" {
		return g.gitLog(dirs, plumbing.ZeroHash)
	}

	repo, err := g.open(".")
	if err != nil {
		return nil, err
	}

	tagRef, err := repo.Tag(tag)
	if err != nil {
		return nil, err
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
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported tag type: %T", obj)
		}
	case plumbing.ErrObjectNotFound:
		return nil, fmt.Errorf("tag not found: %s", tag)
	default:
		return nil, err
	}

	return g.gitLog(dirs, tagCommit.Hash)
}

func (g *Git) gitLog(dirs []string, since plumbing.Hash) ([]Commit, error) {
	repo, err := g.open(".")
	if err != nil {
		return nil, err
	}

	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commitIter, err := repo.Log(&git.LogOptions{
		From: headRef.Hash(),
	})
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

		if since != plumbing.ZeroHash && commit.Hash == since {
			break
		}

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
			commit.Hash.String(),
			title,
			strings.TrimSpace(body),
		})
	}

	return result, nil
}
