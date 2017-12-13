# svu

Semantic Version Util is a tool to increase the major/minor/patch parts
of semantic versions.

The idea is to use it in automated scripts, so you can have in your
makefile something like:

```Makefile
major:
  tag=$(svu major)
  git tag $tag
  git push $tag
  goreleaser --rm-dist
```

### Install

```sh
go get github.com/caarlos0/svu
```
