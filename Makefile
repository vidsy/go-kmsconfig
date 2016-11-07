BRANCH = "master"
VERSION = $(shell cat ./VERSION)

check-version:
	@echo "Checking if value of VERSION file exists as Git tag..."
	git fetch
	(! git rev-list ${VERSION})

push-tag:
	git checkout ${BRANCH}
	git pull origin ${BRANCH}
	git tag ${VERSION}
	git push origin ${BRANCH} --tags
