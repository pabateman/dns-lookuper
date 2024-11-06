export GO111MODULE ?= on
export CGO_ENABLED ?= 0

PROJECT   ?= dns-lookuper
REPO      ?= github.com/pabateman
REPOPATH  ?= $(REPO)/$(PROJECT)
COMMIT    := $(shell git rev-parse HEAD)
VERSION   ?= $(shell git describe --always --tags)
GOPATH    ?= $(shell go env GOPATH)
ARTIFACT_VERSION ?= local

NPROCS = $(shell grep -c 'processor' /proc/cpuinfo)
MAKEFLAGS += -j$(NPROCS)

BUILDDIR   := $(shell pwd)/out
PLATFORMS  ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64
DISTFILE   := $(BUILDDIR)/$(PROJECT)-$(VERSION)-source.tar.gz
COMPRESS := gzip --best -k -c

define doUPX
	upx -9q --force-macos $@
endef

.PHONY: help
help:
	@echo 'Valid make targets:'
	@echo '  - all:      build binaries for all supported platforms'
	@echo '  - build:    build binaries for all supported platforms'
	@echo '  - clean:    clean up build directory'
	@echo '  - deploy:   build artifacts for a new deployment'
	@echo '  - dev:      build the binary for the current platform'
	@echo '  - dist:     create a tar archive of the source code'
	@echo '  - lint:     run golangci-lint'
	@echo '  - help:     print this help'

$(BUILDDIR):
	mkdir -p "$@"

.PHONY: all
all: lint build deploy

.PHONY: gitignore
gitignore:
	@echo $(PROJECT)
	@echo $(notdir $(BUILDDIR))/
	@echo !$(notdir $(BUILDDIR))/.gitkeep

.PHONY: lint
lint:
	golangci-lint run \
		--timeout=3m

.INTERMEDIATE: $(DISTFILE:.gz=)
$(DISTFILE:.gz=): $(BUILDDIR)
	git archive --prefix="$(PROJECT)-$(VERSION)/" --format=tar HEAD > "$@"

.PHONY: dist
dist: $(DISTFILE)


.PHONY: dev
dev:
	go build -o $(PROJECT) cmd/$(PROJECT)/main.go

.PHONY: image
image:
	docker build -t $(notdir $(REPO))/$(PROJECT):$(ARTIFACT_VERSION) .

.PHONY: test
test:
	go test ./... -cover -coverpkg=./... -v

PLATFORM_TARGETS := $(PLATFORMS)

$(PLATFORM_TARGETS):
	$(eval OSARCH := $(subst /, ,$@))
	$(eval OS := $(firstword $(OSARCH)))
	$(eval ARCH := $(lastword $(OSARCH)))
	GOARCH=$(ARCH) GOOS=$(OS) \
	go build \
		-trimpath \
		-o $(BUILDDIR)/$(PROJECT)-$(OS)-$(ARCH) \
		-ldflags="\
			-X main.Version=$(ARTIFACT_VERSION) \
			-X main.GoVersion=$(shell go version | cut -d " " -f 3) \
			-X main.Compiler=$(shell go env CC)                     \
			-X main.Platform=$(shell go env GOOS)/$(shell go env GOARCH)  \
			" \
		cmd/$(PROJECT)/main.go

.PHONY: build
build: $(BUILDDIR) $(PLATFORM_TARGETS)

.PRECIOUS: %.gz
%.gz: %
	$(COMPRESS) "$<" > "$@"

%.tar: %
	cp LICENSE $(BUILDDIR)
	tar cf "$@" -C $(BUILDDIR) LICENSE $(patsubst $(BUILDDIR)/%,%,$^)
	$(RM) $(BUILDDIR)/LICENSE $(patsubst %.tar, %, $@)

%.sha256: %
	sha256sum  $< > $@

$(foreach platform, $(PLATFORM_TARGETS), $(platform)/archive): SUFFIX = $(subst /,-,$(patsubst %/archive,%,$@))
$(foreach platform, $(PLATFORM_TARGETS), $(platform)/archive): $(BUILDDIR)
	$(MAKE) $(BUILDDIR)/$(PROJECT)-$(SUFFIX).tar.gz.sha256

.PHONY: deploy
deploy: $(foreach platform, $(PLATFORMS), $(platform)/archive)

.PHONY: clean
clean:
	$(RM) -r $(BUILDDIR)/* $(PROJECT)

$(foreach platform, $(PLATFORMS), $(BUILDDIR)/$(PROJECT)-$(firstword $(subst /, ,$(platform)))-$(lastword $(subst /, ,$(platform)))):
	$(eval basename := $(notdir $@))
	$(MAKE) $(subst -,/, $(patsubst $(PROJECT)-%, %, $(basename)))
