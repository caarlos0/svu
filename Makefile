major:
	git tag $$(svu major)
	git push --tags
	goreleaser --clean
.PHONY: major

minor:
	git tag $$(svu minor)
	git push --tags
	goreleaser --clean
.PHONY: minor

patch:
	git tag $$(svu patch)
	git push --tags
	goreleaser --clean
.PHONY: patch

.DEFAULT_GOAL := patch
