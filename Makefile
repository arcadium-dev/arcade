# Copyright 2021-2022 arcadium.dev <info@arcadium.dev>
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

export SHELL := /bin/bash

export app := arcade

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
export date := $(shell date -u --iso-8601='seconds')

# ldflags are the go linker flags we pass to the go command.
#   -s    Omit the symbol table and debug information.
#   -w    Omit the DWARF symbol table.
#   -X importpath.name=value
#         Set the value of the string variable in import path named name to
#         value.  This is only effective if the variable is declared in the
#         source code either uninitialized or initialized to a constant string
#         expression.
export ldflags := -s -w -X main.version=$(version) -X main.branch=$(branch) -X main.commit=$(shasum) -X main.date=$(date)

# ____ all __________________________________________________________________

.PHONY: all

all: build test lint containers

# ____ lint __________________________________________________________________

.PHONY: fmt tudy lint

fmt:
	@printf "\nRunning go fmt...\n"
	go fmt ./...

tidy:
	@printf "\nRunning go mod tidy...\n"
	go mod tidy

lint: fmt tidy
	@printf "\nChecking for changed files...\n"
	git status --porcelain
	@printf "\n"
	@if [[ "$${CI}" == "true" ]]; then $$(exit $$(git status --porcelain | wc -l)); fi

# ____ test __________________________________________________________________

.PHONY: unit_test migrate_test integration_test test

unit_test:
	@printf "\nRunning go test...\n"
	go test -cover -race ./...

dev_init: containers
	@bin/dev init networks
	@bin/dev init certs
	@bin/dev init db

migrate_test: dev_init
	@bin/dev init migrate up
	@bin/dev init migrate down -all

integration_test: migrate_test
	@bin/dev init migrate up
	@bin/dev start cockroach assets
	@bin/dev ps
	@bin/integration_test; rc=$$?; \
	  if [ $$rc -ne 0 ]; then bin/dev logs haproxy infra; fi;\
	  bin/dev stop;\
	  exit $$rc

test: unit_test integration_test

# ____ build _________________________________________________________________

.PHONY: build assets version

build: assets

assets:
	@printf "\nBuilding assets...\n"
	CGO_ENABLED=0 go build -ldflags "$(ldflags)" -o ./dist/assets ./cmd/assets

version:
	@printf "\nVersion: $(version)\n"

# ____ container artifacts ___________________________________________________

.PHONY: containers push_containers

containers:
	 make -C dockerfiles all

migrate_container:
	make -C dockerfiles migrate_container

push_containers:
	make -C dockerfiles push_containers

# ____ clean artifacts _______________________________________________________

.PHONY: clean

clean:
	@printf "\nClean...\n"
		@bin/dev stop
		@bin/dev clean
		-rm -rf dist
