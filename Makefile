major:
	git tag $$(svu major)
	git push --tags
	goreleaser --rm-dist
.PHONY: major

minor:
	git tag $$(svu minor)
	git push --tags
	goreleaser --rm-dist
.PHONY: minor

patch:
	git tag $$(svu patch)
	git push --tags
	goreleaser --rm-dist
.PHONY: patch

.DEFAULT_GOAL := patch
