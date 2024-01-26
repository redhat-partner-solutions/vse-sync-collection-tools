.PHONY:
	install-tools
	build
	build-image
	lint

# Get default value of $GOBIN if not explicitly set
GO_PATH=$(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
	GOBIN=${GO_PATH}/bin
else
	GOBIN=$(shell go env GOBIN)
endif

# Install build tools and other required software.
install-tools:
	go install github.com/onsi/ginkgo/v2/ginkgo

# Build test binary
build:
	PATH=${PATH}:${GOBIN} go build --race

build-image:
	podman build -t synctest:custom -f Containerfile

lint:
	golangci-lint run --timeout 5m0s --fix
