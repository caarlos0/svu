<p align="center">
  <img alt="svu Logo" src="https://becker.software/svu.png" height="300" />
  <p align="center">semantic version utility</p>
</p>

<hr>

<p align="center">
<a href="https://github.com/caarlos0/svu/releases/latest"><img src="https://img.shields.io/github/release/caarlos0/svu.svg?style=for-the-badge" alt="Release"></a>
<a href="/LICENSE.md"><img src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge" alt="Software License"></a>
<a href="https://github.com/caarlos0/svu/actions?workflow=build"><img src="https://img.shields.io/github/actions/workflow/status/caarlos0/svu/build.yml?style=for-the-badge&branch=main" alt="Build status"></a>
<a href="http://godoc.org/github.com/caarlos0/svu/v3"><img src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge" alt="Go Doc"></a>
<a href="https://goreportcard.com/report/github.com/caarlos0/svu/v3"><img src="https://goreportcard.com/badge/github.com/caarlos0/svu/v3?style=for-the-badge" alt="GoReportCard"></a>
<a href="https://conventionalcommits.org"><img src="https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge" alt="Conventional Commits"></a>
</p>


semantic version utility (svu) is a small helper for release scripts and workflows.

It provides utility commands and functions to increase specific portions of the version.
It can also figure the next version out automatically by looking through the git history.

> [!TIP]
> Read [the spec][Semver] for more information.

## usage

Check `svu --help` for the list of sub-commands and flags.

### `next`, `n`

This is probably the command you'll use the most.

It checks your `git log`, and automatically increases and returns the new
version based on this table:

| Commit message                                                                         | Tag increase |
| -------------------------------------------------------------------------------------- | ------------ |
| `chore: foo`                                                                           | Nothing      |
| `fix: fixed something`                                                                 | Patch        |
| `feat: added new button to do X`                                                       | Minor        |
| `fix: fixed thing xyz`<br><br>`BREAKING CHANGE: this will break users because of blah` | Major        |
| `fix!: fixed something`                                                                | Major        |
| `feat!: added blah`                                                                    | Major        |

> [!TIP]
> You can create an alias to create tags automatically:
>
> ```bash
> alias gtn='git tag $(svu next)'
> ```

## configuration

Every flag option can also be set in a `.svu.yml` in the current
directory/repository root folder, for example:

```yaml
tag.prefix: ""
always: true
v0: true
```

Names are the same as the flags themselves.

## install

[![Packaging status](https://repology.org/badge/vertical-allrepos/svu.svg)](https://repology.org/project/svu/versions)

<details>
  <summary>macOS</summary>

```bash
brew install --cask caarlos0/tap/svu
```

</details>

<details>
  <summary>linux/apt</summary>

```bash
echo 'deb [trusted=yes] https://apt.fury.io/caarlos0/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install svu
```

</details>

<details>
  <summary>linux/yum</summary>

```bash
echo '[caarlos0]
name=caarlos0
baseurl=https://yum.fury.io/caarlos0/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install svu
```

</details>

<details>
  <summary>docker</summary>

```bash
docker run --rm -v $PWD:/tmp --workdir /tmp ghcr.io/caarlos0/svu --help
```

</details>

<details>
  <summary><code>go install</code></summary>

```bash
go install github.com/caarlos0/svu/v3@latest
```

</details>

<details>
  <summary>manually</summary>

Or download one from the [releases tab](https://github.com/caarlos0/svu/releases) and install manually.

</details>

## stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/svu.svg?variant=adaptive)](https://starchart.cc/caarlos0/svu)

[Semver]: https://semver.org

---

Logo art and concept by [@carinebecker](https://github.com/carinebecker).
