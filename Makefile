BRANCH = "master"
PACKAGES ?= "$(shell glide nv)"
PWD = $(shell pwd)
REPONAME ?= "go-kmsconfig"
S3_BUCKET ?= "go-kmsconfig.live.vidsy.co"
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

lint-ci:
	@docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:${PWD} -v ~/.circleci/:/root/.circleci --workdir ${PWD} circleci/picard:latest circleci config -c .circleci/config.yml validate

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

run:
	@if test -z $(path); then echo "Please specify a config file path"; exit 1; fi
	@if test -z $(node); then echo "Please specify a config node"; exit 1; fi
	@docker run --rm -v $(path):/config -e AWS_ENV=${AWS_ENV} vidsyhq/${REPONAME}:latest -path /config -node $(node)

test:
	@go test "${PACKAGES}" -cover

vet:
	@go vet "${PACKAGES}"
