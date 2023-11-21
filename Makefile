export app := arcade

export SHELL := /bin/bash

go_version := 1.21
ifeq ($(shell uname -m),arm64)
  arch := arm64
else
  arch := amd64
endif

# sha_len is the length of the sha sum used with the version and the sha
sha_len := 7

# version is the version of the current branch. For code that matches a
# released version we want the exact version match, i.e. v1.0.0. For code that
# is part of work in progress we want a version that denotes a path to a
# release version, i.e. v1.0.0-5-g07a65db-dirty, where the closest release is
# v1.0.0, the 5 denotes that the code is 5 commits ahead of the release,
# g07a65db is the git sha of the latest commit, and dirty denotes that there
# are uncommitted changes to the code.
export version := $(shell git describe --tags --dirty --abbrev=$(sha_len))

# branch is the name of the current branch
export branch ?= $(shell git rev-parse --abbrev-ref HEAD)

# commit is the shasum of the latest commit
export commit := $(shell git rev-parse HEAD)

# date is the date of the build
export date := $(shell date -u -Iseconds)

# ldflags are the go linker flags we pass to the go command.
#   -s    Omit the symbol table and debug information.
#   -w    Omit the DWARF symbol table.
#   -X importpath.name=value
#         Set the value of the string variable in import path named name to
#         value.  This is only effective if the variable is declared in the
#         source code either uninitialized or initialized to a constant string
#         expression.
export ldflags := -s -w \
	-X main.Version=$(version) \
	-X main.Branch=$(branch) \
	-X main.Commit=$(commit) \
	-X main.Date=$(date)

# ____ all __________________________________________________________________

.PHONY: all

all: test lint

# ____ lint __________________________________________________________________

.PHONY: fmt tidy vet staticcheck vuln lint

fmt:
	@printf "\nRunning go fmt...\n"
	go fmt ./...

tidy:
	@printf "\nRunning go mod tidy...\n"
	go mod tidy

vet:
	@printf "\nRunning go vet...\n"
	go vet ./...

staticcheck:
	@if [[ ! -x "$$(go env GOPATH)/bin/staticcheck" ]]; then \
		printf "\nInstalling staticcheck...\n"; \
		go install "honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi
	@printf "\nRunning staticcheck...\n"
	$$(go env GOPATH)/bin/staticcheck ./...

vuln:
	@if [[ ! -x "$$(go env GOPATH)/bin/govulncheck" ]]; then \
		printf "\nInstalling govulncheck...\n"; \
		go install "golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi
	@printf "\nRunning govulncheck...\n"
	$$(go env GOPATH)/bin/govulncheck ./...

lint: fmt tidy vet staticcheck vuln
	@printf "\nChecking for changed files...\n"
	git status --porcelain
	@printf "\n"
	@if [[ "$${CI}" == "true" ]]; then $$(exit $$(git status --porcelain | wc -l)); fi

# ____ test __________________________________________________________________

.PHONY: unit_test test

unit_test:
	@printf "\nRunning go test...\n"
	go test -cover -race ./...

test: unit_test

integration_test_build:
	dev nuke && make dev-images && dev init

esc := \033
clear := $(esc)[0;39m
yellow := $(esc)[1;33m

integration_test:
	@mkdir -m 0777 -p ./asset/test/coverage
	dev start
	@sleep 3; echo -e "\n$(yellow)Running Assets Integration Tests$(clear)"
	@-go test ./asset/test
	@echo -e "\n$(yellow)Assets Coverage$(clear)"
	@go tool covdata percent -i=./asset/test/coverage
	@echo ""
	@dev stop
	@rm -rf ./asset/test/coverage

# ____ docs __________________________________________________________________

.PHONY: docs
docs:
	@if [[ ! -x "$$(go env GOPATH)/bin/swagger" ]]; then \
		printf "\nInstalling go-swagger...\n"; \
		go install "github.com/go-swagger/go-swagger/cmd/swagger@latest"; \
	fi
	@printf "\nRunning swagger...\n"
	$$(go env GOPATH)/bin/swagger generate spec -o ./docs/swagger.json

# ____ image artifacts  __________________________________________________

.PHONY: images assets assets-migrate mkcert curl

export buildargs :=

images:
	make -C dockerfiles all

dev-images: buildargs := -cover
dev-images:
	make -C dockerfiles all

assets assets-migrate mkcert curl:
	make -C dockerfiles $@

# ____ clean artifacts _______________________________________________________

.PHONY: clean

clean:
	@printf "\nClean...\n"
	-rm -rf ./dist ./test/seed
	-rm -rf ./asset/test/coverage
	-go clean -testcache -cache
	-rm -f $$(go env GOPATH)/bin/staticcheck
	-rm -f $$(go env GOPATH)/bin/govulncheck
	-rm -f $$(go env GOPATH)/bin/swagger
