# svu

[![Build Status](https://img.shields.io/github/actions/workflow/status/caarlos0/svu/build.yml?style=for-the-badge)](https://github.com/caarlos0/svu/actions?workflow=build)

Semantic Version Util is a tool to manage semantic versions at ease!

You can print the current version, increase patch/minor/major manually or just
get the next tag based on your git log!

## example usage

### `svu`

Based on the log between the latest tag and `HEAD`, prints the next tag.

> aliases: `svu next` and `svu n`

```bash
$ svu
v1.3.0
```

#### commit messages vs what they do

| Commit message                                                                         | Tag increase |
| -------------------------------------------------------------------------------------- | ------------ |
| `fix: fixed something`                                                                 | Patch        |
| `feat: added new button to do X`                                                       | Minor        |
| `fix: fixed thing xyz`<br><br>`BREAKING CHANGE: this will break users because of blah` | Major        |
| `fix!: fixed something`                                                                | Major        |
| `feat!: added blah`                                                                    | Major        |
| `chore: foo`                                                                           | Nothing      |

### `svu current`

Prints the current tag with no changes.

> alias: `svu c`

**Examples:**

```bash
$ svu current
v1.2.3

$ svu current
v1.2.3-alpha.1+22
```

### `svu major`

Increases the major of the latest tag and prints it.
As per the [Semver][] spec, it'll also clear the `pre-release` and `build`
identifiers are cleaned up.

**Examples:**

```bash
$ svu current
v1.2.3-alpha.2+123

$ svu major
v2.0.0

$ svu major --pre-release alpha.3 --build 243
v2.0.0-alpha.3+243
```

### `svu minor`

Increases the minor of the latest tag and prints it.
As per the [Semver][] spec, it'll also clear the `pre-release` and `build`
identifiers are cleaned up.

> alias: `svu m`

**Examples:**

```bash
$ svu current
v1.2.3-alpha.2+123

$ svu minor
v1.3.0

$ svu minor --pre-release alpha.3 --build 243
v1.3.0-alpha.3+243
```

### `svu patch`

Increases the patch of the latest tag and prints it.
As per the [Semver][] spec, if the version has a `pre-release` or `build`
identifier, they'll be cleaned up and no patch increment will be made.
You can force a patch increment by using `svu next --force-patch-increment`.

> alias: `svu p`

**Examples:**

```bash
$ svu current
v1.2.3-alpha.2+123

$ svu patch
v1.2.3

$ svu patch --pre-release alpha.3 --build 243
v1.2.3-alpha.3+243

$ svu next --force-patch-increment
v1.2.4
```

### `svu prerelease`

Increases the pre-release of the latest tag and prints it. If a `pre-release`
identifier is passed in and it differs from the current pre-release tag that
the identifier passed in will be used. If the current tag is not a pre-release
tag then passing in `--pre-release` is required.

> alias: `svu pr`

**Examples:**

```bash
$ svu current
v1.2.3-alpha.2+123

$ svu prerelease
v1.2.3-alpha.3

$ svu prerelease --pre-release alpha.33 --build 243
v1.2.3-alpha.33+243
```

## tag mode

By default `svu` will get the latest tag from the current branch. Using the `--tag-mode` flag this behaviour can be altered:

| Flag                        | Description                          | Git command used under the hood                            |
| --------------------------- | ------------------------------------ | ---------------------------------------------------------- |
| `--tag-mode current-branch` | Get latest tag from current branch.  | `git describe --tags --abbrev=0`                           |
| `--tag-mode all-branches`   | Get latest tag across all branches. | `git describe --tags $(git rev-list --tags --max-count=1)` |

## stripping the tag prefix

`--strip-prefix` removes any prefix from the version output.
For example, `v1.2.3` would be output as `1.2.3`.

The default prefix is `v`, however a custom prefix can be supplied using `--prefix`.
So for `--prefix=foo/v --strip-prefix` and tag `foo/v1.2.3`, the output would be `1.2.3`.

## adding a suffix

You can use `--pre-release` and `--build` to set the respective [Semver][]
identifiers to the resulting version.

## force patch version increment

Setting the `--force-patch-increment` flag forces a patch version increment regardless of the commit message content.

**Example:**

```bash
svu next --force-patch-increment
```

## creating tags

The idea is that `svu` will just print things, so it's safe to run at any time.

You can create tags by wrapping it in an alias. For example, I have one like
this:

```bash
alias gtn='git tag $(svu next)'
```

So, whenever I want to create a tag, I just run `gtn`.

## install

[![Packaging status](https://repology.org/badge/vertical-allrepos/svu.svg)](https://repology.org/project/svu/versions)

### macOS

```bash
brew install caarlos0/tap/svu
```

### linux

#### apt

```bash
echo 'deb [trusted=yes] https://apt.fury.io/caarlos0/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install svu
```

#### yum

```bash
echo '[caarlos0]
name=caarlos0
baseurl=https://yum.fury.io/caarlos0/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install svu
```

### docker

```bash
docker run --rm -v $PWD:/tmp --workdir /tmp ghcr.io/caarlos0/svu --help
```

### Using `go install`

Make sure that `$GOPATH/bin` is in your `$PATH`, because that's where this gets
installed:

```bash
go install github.com/caarlos0/svu@latest
```

#### manually

Or download one from the [releases tab](https://github.com/caarlos0/svu/releases) and install manually.

## use as library

You can use `svu` as a library without the need to install the binary. For example to use it from a magefile:

```go
//go:build mage
// +build mage

package main

import (
	"github.com/caarlos0/svu/pkg/svu"
	"github.com/magefile/mage/sh"
	"strings"
)

// Tag the current commit with the proper next semver.
func Version() error {
	v, err := svu.Next()
	if err != nil {
		return err
	}
	return sh.RunV("git", "tag", "-a", v, "-m", strings.Replace(v, "v", "Version ", 1))
}
```

### commands

All commands are available with a function named accordingly:

- `svu.Next()`
- `svu.Current()`
- `svu.Major()`
- `svu.Minor()`
- `svu.Patch()`
- `svu.PreRelease()`

### options

All flags have a matching option function to configure the previous commands beyond their default bahavior:

- `svu.Current(svu.WithPattern("p*"))`
- `svu.Next(svu.WithPrefix("ver"))`
- `svu.Major(svu.StripPrefix())`
- `svu.Minor(svu.WithPreRelease("pre"))`
- `svu.Patch(svu.WithBuild("3"))`
- `svu.Next(svu.WithDirectory("internal"))`
- `svu.Next(svu.WithTagMode(svu.AllBranches))` or `svu.Next(svu.ForAllBranches())`
- `svu.Next(svu.WithTagMode(svu.CurrentBranch))` or `svu.Next(svu.ForCurrentBranch())`
- `svu.Next(svu.ForcePatchIncrement())`

Or multiple options:

- `svu.Next(svu.WithPreRelease("pre"), svu.WithBuild("3"), svu.StripPrefix())`
- `svu.PreRelease(svu.WithPreRelease("alpha.33"), svu.WithBuild("243"))`

## stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/svu.svg)](https://starchart.cc/caarlos0/svu)

[Semver]: https://semver.org
