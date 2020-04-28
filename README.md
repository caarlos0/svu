# svu

Semantic Version Util is a tool to manage semantic versions at ease!

You can print the current version, increase patch/minor/major manually or just
get the next tag based on your git log!

## Example usage:

##### `svu`

> alias: `svu next` and `svu n`

```sh
$ svu
v1.3.0
```

##### `svu current`

> alias: `svu c`

```sh
$ svu current
v1.2.3
```

##### `svu major`

```sh
$ svu major
v2.0.0
```

##### `svu minor`

> alias: `svu m`

```sh
$ svu minor
v1.3.0
```

##### `svu patch`

> alias: `svu p`

```sh
$ svu patch
v1.2.4
```

## Install

```sh
go get github.com/caarlos0/svu
```

or

```sh
brew install caarlos0/tap/svu
```

or

```sh
curl -sfL https://install.goreleaser.com/github.com/caarlos0/svu.sh | bash -s -- -b /usr/local/bin
```

Or download one from the releases tab and install manually.
