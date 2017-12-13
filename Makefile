major:
	git tag $(svu major)
	git push --tags
	goreleaser --rm-dist

minor:
	git tag $(svu minor)
	git push --tags
	goreleaser --rm-dist

patch:
	git tag $(svu patch)
	git push --tags
	goreleaser --rm-dist
