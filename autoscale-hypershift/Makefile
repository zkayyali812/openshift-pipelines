all:
	@echo "Manual launch commands:"
	@echo "  make params      # Display configuration options"
	@echo "  make options     # Display the override options.env"
	@echo "  make scale-up    # Manually scale-up hypershiftdeployments and enable autoscaling"
	@echo "  make scale-down  # Manually scale-down hypershiftdeployments to 1 replica and disable autoscaling"
	@echo "  make cronjobs    # Deploys the CronJobs"
	@echo "  make roles       # Deploys the ClusterRole and ClusterRoleBinding"
	@echo ""
	@echo "Development commands"
	@echo "  make compile    # Compile code"
	@echo "  make build      # Build image"
	@echo "  make push       # Build and push the image used by manual and cronjobs"
	@echo "  make tag-latest # Pushes the latest tag for the image"
	@echo ""
	@echo "Clean up:"
	@echo "  make clean          # Deletes image from registry"
	@echo "  make clean-cronjobs # Deletes the CronJobs"
	@echo "  make clean-roles    # Deletes the ClusterRole and ClusterRoleBinding"

REPO_URL ?= quay.io/acm-sre
VERSION ?= 0.0.1
options.env:
	touch options.env

options:
	@cat ./options.env
	@echo ""

params:
	oc process -f templates/cronjobs.yaml --parameters

clean: checks
	docker image rm ${REPO_URL}/autoscale-hypershift:${VERSION}

scale-up: options.env
	oc process -f templates/job-scale-up.yaml --param-file options.env --ignore-unknown-parameters=true | oc apply -f -

scale-down: options.env
	oc process -f templates/job-scale-down.yaml --param-file options.env --ignore-unknown-parameters=true  | oc apply -f -

cronjobs: options.env
	oc process -f templates/cronjobs.yaml --param-file options.env --ignore-unknown-parameters=true  | oc apply -f -

clean-cronjobs: options.env
	oc process -f templates/cronjobs.yaml --param-file options.env --ignore-unknown-parameters=true  | oc delete -f -

roles: options.env
	oc process -f templates/roles.yaml --param-file options.env --ignore-unknown-parameters=true | oc apply -f -

clean-roles: options.env
	oc process -f templates/roles.yaml --param-file options.env --ignore-unknown-parameters=true | oc delete -f -

compile:
	go mod tidy
	go build -o action .

build:
	docker build -f Dockerfile . -t ${REPO_URL}/autoscale-hypershift:${VERSION}

push: build
	docker push ${REPO_URL}/autoscale-hypershift:${VERSION}

tag-latest: push
	docker tag ${REPO_URL}/autoscale-hypershift:${VERSION} ${REPO_URL}/autoscale-hypershift:latest
	docker push ${REPO_URL}/autoscale-hypershift:latest

scale-up-local:
	 TAKE_ACTION="ScaleUp" go run .

scale-down-local:
	 TAKE_ACTION="ScaleDown" go run .

