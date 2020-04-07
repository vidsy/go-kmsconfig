BRANCH = "master"
PACKAGES ?= "$(shell glide nv)"
PWD = $(shell pwd)
REPONAME ?= "go-kmsconfig"
VERSION = $(shell cat ./VERSION)

build-image:
	@docker build -t vidsyhq/${REPONAME} .

check-version:
	@echo "Checking if value of VERSION file exists as Git tag..."
	(! git rev-list ${VERSION})

install:
	@echo "=> Install dependencies"
	@GO111MODULE=on go mod download

lint-ci:
	@docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:${PWD} -v ~/.circleci/:/root/.circleci --workdir ${PWD} circleci/picard:latest circleci config -c .circleci/config.yml validate

push-tag:
	@echo "=> New tag version: v${VERSION}"
	git checkout ${BRANCH}
	git pull origin ${BRANCH}
	git tag v${VERSION}
	git push origin v${VERSION}

push-to-registry:
	@docker login -e ${DOCKER_EMAIL} -u ${DOCKER_USER} -p ${DOCKER_PASS}
	@docker tag vidsyhq/${REPONAME}:latest vidsyhq/${REPONAME}:${CIRCLE_TAG}
	@docker push vidsyhq/${REPONAME}:${CIRCLE_TAG}
	@docker push vidsyhq/${REPONAME}

release:
	rm -rf dist
	@GITHUB_TOKEN=${VIDSY_GOBOT_GITHUB_TOKEN} VERSION=${VERSION} BUILD_TIME=%${BUILD_TIME} goreleaser

run:
	@if test -z $(path); then echo "Please specify a config file path"; exit 1; fi
	@if test -z $(node); then echo "Please specify a config node"; exit 1; fi
	@docker run --rm -v $(path):/config -e AWS_ENV=${AWS_ENV} vidsyhq/${REPONAME}:latest -path /config -node $(node)

test:
	@GO111MODULE=on go test "${PACKAGES}" -cover

vet:
	@GO111MODULE=on go vet "${PACKAGES}"
