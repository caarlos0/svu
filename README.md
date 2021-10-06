# svu

[![Build Status](https://img.shields.io/github/workflow/status/caarlos0/svu/build?style=for-the-badge)](https://github.com/caarlos0/svu/actions?workflow=build)

Semantic Version Util is a tool to manage semantic versions at ease!

You can print the current version, increase patch/minor/major manually or just
get the next tag based on your git log!

## example usage

### `svu`

Based on the log between the latest tag and `HEAD`, prints the next tag.

> aliases: `svu next` and `svu n`

```sh
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

Prints the latest tag.

> alias: `svu c`

```sh
$ svu current
v1.2.3
```

### `svu major`

Increases the major of the latest tag and prints it.

```sh
$ svu major
v2.0.0
```

### `svu minor`

Increases the minor of the latest tag and prints it.

> alias: `svu m`

```sh
$ svu minor
v1.3.0
```

### `svu patch`

Increases the patch of the latest tag and prints it.

> alias: `svu p`

```sh
$ svu patch
v1.2.4
```

## tag mode

By default `svu` will get the latest tag from the current branch. Using the `--tag-mode` flag this behaviour can be altered:

| Flag                        | Description                          | Git command used under the hood                            |
| --------------------------- | ------------------------------------ | ---------------------------------------------------------- |
| `--tag-mode current-branch` | Get latest tag from current branch.  | `git describe --tags --abbrev=0`                           |
| `--tag-mode all-branches`   | Get latest tag across all branches. | `git describe --tags $(git rev-list --tags --max-count=1)` |

## discarding pre-release and build metadata

To discard [pre-release](https://semver.org/#spec-item-9) and/or [build metadata](https://semver.org/#spec-item-10) information you can run your command of choice with the following flags:

| Flag               | Description                              | Example                                  |
| ------------------ | ---------------------------------------- | ---------------------------------------- |
| `--no-metadata`    | Discards pre-release and build metadata. | `v1.0.0-alpha+build.f902daf` -> `v1.0.0` |
| `--no-pre-release` | Discards pre-release metadata.           | `v1.0.0-alpha` -> `v1.0.0`               |
| `--no-build`       | Discards build metadata.                 | `v1.0.0+build.f902daf` -> `v1.0.0`       |

## stripping the tag prefix

`--strip-prefix` removes any prefix from the version output.
For example, `v1.2.3` would be output as `1.2.3`.

The default prefix is `v`, however a custom prefix can be supplied using `--prefix`.
So for `--prefix=foo/v --strip-prefix` and tag `foo/v1.2.3`, the output would be `1.2.3`.

## adding a suffix

`--suffix` can be used to set a custom suffix for the tag, e.g. build metadata or prelease.

## force patch version increment

Setting the `--force-patch-increment` flag forces a patch version increment regardless of the commit message content.

## creating tags

The idea is that `svu` will just print things, so it's safe to run at any time.

You can create tags by wrapping it in an alias. For example, I have one like
this:

```bash
alias gtn='git tag $(svu next)'
```

So, whenever I want to create a tag, I just run `gtn`.

## install

### macOS

```sh
brew install caarlos0/tap/svu
```

### linux

#### apt

```sh
echo 'deb [trusted=yes] https://apt.fury.io/caarlos0/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install svu
```

#### yum

```sh
echo '[caarlos0]
name=caarlos0
baseurl=https://yum.fury.io/caarlos0/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install svu
```

### docker

```sh
docker run --rm -v $PWD:/tmp --workdir /tmp ghcr.io/caarlos0/svu --help
```

#### manually

Or download one from the [releases tab](https://github.com/caarlos0/svu/releases) and install manually.

## stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/svu.svg)](https://starchart.cc/caarlos0/svu)
