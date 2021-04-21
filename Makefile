# Docker Image Options
DOCKER_REGISTRY		?= quay.io
DOCKER_FORCE_CLEAN	?= true
DOCKER_IMAGE_PREFIX	?= airshipit
DOCKER_IMAGE_TAG	?= latest
DOCKER_TARGET_STAGE	?= release
PUBLISH			?= false

JUMP_HOST_IMAGE_NAME	?= jump-host
SIP_IMAGE_NAME		?= sip

JUMP_HOST_BASE_IMAGE	?= gcr.io/gcp-runtimes/ubuntu_18_0_4
SIP_BASE_IMAGE		?= gcr.io/distroless/static:nonroot

# Image URLs to build/publish images
JUMP_HOST_IMG	?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(JUMP_HOST_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
SIP_IMG		?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(SIP_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

# Produce CRDs that work back to Kubernetes 1.16
CRD_OPTIONS ?= crd:crdVersions=v1

TOOLBINDIR          := tools/bin

# linting
LINTER              := $(TOOLBINDIR)/golangci-lint
LINTER_CONFIG       := .golangci.yaml

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Proxy options
HTTP_PROXY          ?= http://proxy.foo.com:8000
HTTPS_PROXY         ?= https://proxy.foo.com:8000
NO_PROXY            ?= localhost,127.0.0.1,.svc.cluster.local
USE_PROXY           ?= false

DOCKER_CMD_FLAGS    ?=

# Docker proxy flags
DOCKER_PROXY_FLAGS  := --build-arg http_proxy=$(HTTP_PROXY)
DOCKER_PROXY_FLAGS  += --build-arg https_proxy=$(HTTPS_PROXY)
DOCKER_PROXY_FLAGS  += --build-arg HTTP_PROXY=$(HTTP_PROXY)
DOCKER_PROXY_FLAGS  += --build-arg HTTPS_PROXY=$(HTTPS_PROXY)
DOCKER_PROXY_FLAGS  += --build-arg no_proxy=$(NO_PROXY)
DOCKER_PROXY_FLAGS  += --build-arg NO_PROXY=$(NO_PROXY)

ifeq ($(USE_PROXY), true)
DOCKER_CMD_FLAGS += $(DOCKER_PROXY_FLAGS)
endif

kubernetes:
	./tools/deployment/install-k8s.sh

all: manager

# Run tests
test: generate fmt vet manifests lint api-docs
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${SIP_IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

images: docker-build-controller docker-build-jump-host

# Build the SIP Docker image
docker-build-controller:
	docker build ${DOCKER_CMD_FLAGS} --build-arg BASE_IMAGE=${SIP_BASE_IMAGE} . -t ${SIP_IMG}

# Build the Jump Host Docker image
docker-build-jump-host:
	docker build ${DOCKER_CMD_FLAGS} -f images/jump-host/Dockerfile --build-arg BASE_IMAGE=${JUMP_HOST_BASE_IMAGE} . -t ${JUMP_HOST_IMG}

docker-push-controller:
	docker push ${SIP_IMG}

docker-push-jump-host:
	docker push ${JUMP_HOST_IMG}

# Generate API reference documentation
api-docs: gen-crd-api-reference-docs
	$(API_REF_GEN) -api-dir=./pkg/api/v1 -config=./hack/api-docs/config.json -template-dir=./hack/api-docs/template -out-file=./docs/api/sipcluster.md

# Find or download gen-crd-api-reference-docs
gen-crd-api-reference-docs:
ifeq (, $(shell which gen-crd-api-reference-docs))
	@{ \
	set -e ;\
	API_REF_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$API_REF_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/ahmetb/gen-crd-api-reference-docs@v0.2.0 ;\
	rm -rf $$API_REF_GEN_TMP_DIR ;\
	}
API_REF_GEN=$(GOBIN)/gen-crd-api-reference-docs
else
API_REF_GEN=$(shell which gen-crd-api-reference-docs)
endif

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

flux-helm-controller:
	kustomize build "github.com/fluxcd/helm-controller/config/default/?ref=v0.8.0" | kubectl apply -f -

.PHONY: lint
lint: $(LINTER)
	@echo "Performing linting step..."
	@./tools/whitespace_linter
	@./$(LINTER) run --config $(LINTER_CONFIG)
	@echo "Linting completed successfully"

$(LINTER): $(TOOLBINDIR)
	./tools/install_linter

$(TOOLBINDIR):
	mkdir -p $(TOOLBINDIR)

.PHONY: check-git-diff
check-git-diff:
	@./tools/git_diff_check
