.DEFAULT_GOAL = test

BIN_NAME = obsidian-tasks-mcp

extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GOOS ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/')
GOARCH ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/')

ifeq ("$(TARGETVARIANT)","")
ifneq ("$(GOARM)","")
TARGETVARIANT := v$(GOARM)
endif
else
ifeq ("$(GOARM)","")
GOARM ?= $(subst v,,$(TARGETVARIANT))
endif
endif

ifeq ("$(CI)","true")
LINT_PROCS ?= 1
else
LINT_PROCS ?= $(shell nproc)
endif

# test with race detector on supported platforms
# windows/amd64 is supported in theory, but in practice it requires a C compiler
race_platforms := 'linux/amd64' 'darwin/amd64' 'darwin/arm64'
ifeq (,$(findstring '$(GOOS)/$(GOARCH)',$(race_platforms)))
export CGO_ENABLED=0
TEST_ARGS=
else
TEST_ARGS=-race
endif

test:
	go test $(TEST_ARGS) -coverprofile=c.out ./...

bench.txt:
	go test -benchmem -run=xxx -bench . ./... | tee $@

bin/$(BIN_NAME)_%v7$(call extension,$(GOOS)): $(shell find . -type f -name "*.go")
	GOOS=$(shell echo $* | cut -f1 -d-) \
	GOARCH=$(shell echo $* | cut -f2 -d- ) \
	GOARM=7 \
	CGO_ENABLED=0 \
		go build $(BUILD_ARGS) -o $@ ./

bin/$(BIN_NAME)_windows-%.exe: $(shell find . -type f -name "*.go")
	GOOS=windows \
	GOARCH=$* \
	GOARM= \
		CGO_ENABLED=0 \
		go build $(BUILD_ARGS) -o $@ ./

bin/$(BIN_NAME)_%$(TARGETVARIANT)$(call extension,$(GOOS)): $(shell find . -type f -name "*.go")
	GOOS=$(shell echo $* | cut -f1 -d-) \
	GOARCH=$(shell echo $* | cut -f2 -d- ) \
	GOARM=$(GOARM) \
	CGO_ENABLED=0 \
		go build $(BUILD_ARGS) -o $@ ./

bin/$(BIN_NAME): bin/$(BIN_NAME)_$(GOOS)-$(GOARCH)
	cp $< $@

lint:
	@golangci-lint run --verbose --max-same-issues=0 --max-issues-per-linter=0

ci-lint:
	@golangci-lint run --verbose --max-same-issues=0 --max-issues-per-linter=0 --out-format=github-actions

.PHONY: test lint ci-lint
.DELETE_ON_ERROR:
.SECONDARY:
