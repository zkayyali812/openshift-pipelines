REPO_URL ?= quay.io/acm-sre
VERSION ?= 2.5.0
IMG ?= kcp-ns-cleanup

options.env:
	touch options.env

options:
	@cat ./options.env
	@echo ""

build:
	docker build -f Dockerfile . -t ${REPO_URL}/${IMG}:${VERSION}

push: build
	docker push ${REPO_URL}/${IMG}:${VERSION}

tag-latest: push
	docker tag ${REPO_URL}/${IMG}:${VERSION} ${REPO_URL}/${IMG}:latest
	docker push ${REPO_URL}/${IMG}:latest

compile:
	go mod tidy
	go build -o kcp-ns-cleanup .

build:
	docker build -f Dockerfile . -t ${REPO_URL}/${IMG}:${VERSION}

run-local:
	go run .

roles: options.env
	oc process -f templates/roles.yaml --param-file options.env --ignore-unknown-parameters=true | oc apply -f -

clean-roles: options.env
	oc process -f templates/roles.yaml --param-file options.env --ignore-unknown-parameters=true | oc delete -f -

deploy: options.env
	oc process -f templates/job-kcp-ns-cleanup.yaml --param-file options.env --ignore-unknown-parameters=true | oc apply -f -
