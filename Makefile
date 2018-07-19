BRANCH = "master"
GO_BUILDER_IMAGE ?= "vidsyhq/go-builder"
PACKAGES ?= "$(shell glide nv)"
PATH_BASE ?= "/go/src/github.com/vidsy"
REPONAME ?= "go-kmsconfig"
S3_BUCKET ?= "go-kmsconfig.live.vidsy.co"
TEST_PACKAGES ?= "./kmsconfig ./cli"
VERSION = $(shell cat ./VERSION)

build-image:
	@docker build -t vidsyhq/${REPONAME} .

check-version:
	@echo "Checking if value of VERSION file exists as Git tag..."
	(! git rev-list ${VERSION})

deploy:
	@echo "Deploying version ${VERSION} to S3"
	aws s3 cp go-kmsconfig s3://${S3_BUCKET}/${VERSION}/go-kmsconfig

install:
	@echo "=> Installing dependencies"
	@dep ensure

push-tag:
	git checkout ${BRANCH}
	git pull origin ${BRANCH}
	git tag ${VERSION}
	git push origin ${BRANCH} --tags

push-to-registry:
	@docker login -e ${DOCKER_EMAIL} -u ${DOCKER_USER} -p ${DOCKER_PASS}
	@docker tag vidsyhq/${REPONAME}:latest vidsyhq/${REPONAME}:${CIRCLE_TAG}
	@docker push vidsyhq/${REPONAME}:${CIRCLE_TAG}
	@docker push vidsyhq/${REPONAME}

test:
	@go test "${PACKAGES}" -cover

vet:
	@go vet "${PACKAGES}"
