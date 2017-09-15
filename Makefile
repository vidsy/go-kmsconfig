BRANCH = "master"
GO_BUILDER_IMAGE ?= "vidsyhq/go-builder"
PATH_BASE ?= "/go/src/github.com/vidsy"
REPONAME ?= "go-kmsconfig"
TEST_PACKAGES ?= "./kmsconfig"

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

test:
	@go test $(shell glide nv) -cover

test-ci:
	@docker run \
	-v "${CURDIR}":${PATH_BASE}/${REPONAME} \
	-w ${PATH_BASE}/${REPONAME} \
	--entrypoint=go \
	${GO_BUILDER_IMAGE} test "${TEST_PACKAGES}" -cover
