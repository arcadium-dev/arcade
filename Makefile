# Copyright 2021-2024 arcadium.dev <info@arcadium.dev>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export app := arcade

export SHELL := /bin/bash

go_version := 1.24
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

all: generate test lint

# ____ generate ______________________________________________________________

.PHONY: generate openapi

openapi_src := \
	./internal/user/server/users_service.gen.go \
	./internal/user/client/users_client.gen.go

openapi:
	@if [[ ! -x "$$(go env GOPATH)/bin/oapi-codegen" ]]; then \
    printf "\nInstalling oapi-codegen...\n"; \
    go get "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"; \
		go mod tidy; \
    go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"; \
  fi

./internal/user/server/users_service.gen.go: ./openapi/users-openapi.yaml ./openapi/users-server.yaml
	@printf "\nRunning oapi-codegen for $@...\n"
	$$(go env GOPATH)/bin/oapi-codegen --config=./openapi/users-server.yaml ./openapi/users-openapi.yaml

./internal/user/client/users_client.gen.go: ./openapi/users-openapi.yaml ./openapi/users-client.yaml
	@printf "\nRunning oapi-codegen for $@...\n"
	$$(go env GOPATH)/bin/oapi-codegen --config=./openapi/users-client.yaml ./openapi/users-openapi.yaml

generate: openapi $(openapi_src)
	@go mod tidy

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
		go get "honnef.co/go/tools/cmd/staticcheck"; \
		go mod tidy; \
		go install "honnef.co/go/tools/cmd/staticcheck"; \
	fi
	@printf "\nRunning staticcheck...\n"
	$$(go env GOPATH)/bin/staticcheck ./...

vuln:
	@if [[ ! -x "$$(go env GOPATH)/bin/govulncheck" ]]; then \
		printf "\nInstalling govulncheck...\n"; \
		go get "golang.org/x/vuln/cmd/govulncheck"; \
		go mod tidy; \
		go install "golang.org/x/vuln/cmd/govulncheck"; \
	fi
	@printf "\nRunning govulncheck...\n"
	$$(go env GOPATH)/bin/govulncheck ./...

lint: fmt tidy vet staticcheck vuln
	@printf "\nChecking for changed files...\n"
	git status --porcelain
	@printf "\n"
	@if [[ "$${CI}" == "true" ]]; then git diff && $$(exit $$(git status --porcelain | wc -l)); fi

# ____ test __________________________________________________________________

.PHONY: unit_test test

unit_test:
	@printf "\nRunning go test...\n"
	go test -cover -race ./...

test: unit_test


.PHONY: integration_test_build assets_integration_test_up users_integration_test_up users_integration_test_up assets_integration_test users_integration_test integration_test

integration_test_build:
	dev nuke && make dev-images && dev init && dev pull telegraf

assets_integration_test_up:
	@mkdir -m 0777 -p ./asset/test/coverage
	@rm -rf ./asset/test/coverage/*

users_integration_test_up:
	@mkdir -m 0777 -p ./user/test/coverage
	@rm -rf ./user/test/coverage/*

integration_test_up: assets_integration_test_up users_integration_test_up
	dev start
	@sleep 5

esc := \033
clear := $(esc)[0;39m
yellow := $(esc)[1;33m

assets_integration_test:
	@echo -e "\n$(yellow)Running Assets Integration Tests$(clear)"
	@-INTEGRATION=1 go test -v --timeout 20s -count=1 ./asset/test

users_integration_test:
	@echo -e "\n$(yellow)Running Users Integration Tests$(clear)"
	@-INTEGRATION=1 go test -v --timeout 20s -count=1 ./user/test

integration_test: integration_test_up assets_integration_test users_integration_test
	dev stop assets users
	@sleep 5
	@echo -e "\n$(yellow)Assets Coverage$(clear)"
	@go tool covdata percent -i=./asset/test/coverage
	@echo -e "\n$(yellow)Users Coverage$(clear)"
	@go tool covdata percent -i=./user/test/coverage

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

.PHONY: images assets users migrate mkcert curl

export buildargs :=

images:
	make -C dockerfiles all

dev-images: buildargs := -cover
dev-images:
	make -C dockerfiles all

assets users migrate mkcert curl:
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
	-rm -f $$(go env GOPATH)/bin/oapi-codegen
