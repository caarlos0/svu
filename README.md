# svu

Semantic Version Util is a tool to increase the major/minor/patch parts
of semantic versions.

The idea is to use it in automated scripts, so you can have in your
makefile something like:

```Makefile
major:
	git tag $$(svu major)
	git push --tags
	goreleaser --rm-dist
```

### Install

```sh
go get github.com/caarlos0/svu
```

or

```sh
brew install caarlos0/tap/svu
```

Or download one from the releases tab and install manually.
