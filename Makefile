# Copyright 2021 arcadium.dev <info@arcadium.dev>
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

app := arcade

# sha_len is the length of the sha sum used with the version and the sha
sha_len := 7

# version is the version of the current branch. For code that matches a
# released version we want the exact version match, i.e. v1.0.0. For
# code that is part of work in progress we want a version that
# denotes a path to a release version, i.e. v1.0.0-5-g07a65db-dirty,
# where the closest release is v1.0.0, the 5 denotes that the code is
# 5 commits ahead of the release, g07a65db is the git sha of
# the latest commit, and dirty denotes that there are uncommitted
# changes to the code.
export version := $(shell git describe --tags --dirty --abbrev=$(sha_len))

# container_version is the version with the initial 'v' removed.
export container_version := $(subst v,,$(version))

# branch is the name of the current branch
export branch ?= $(shell git rev-parse --abbrev-ref HEAD)

# SHASUM is the sha sum of the latest commit
export shasum := $(shell git rev-parse --short=$(sha_len) HEAD)

# date is the date of the build
export date := $(shell date -u --iso-8601='seconds')

# ldflags are the go linker flags we pass to the go command.
#   -s    Omit the symbol table and debug information.
#   -w    Omit the DWARF symbol table.
#   -X importpath.name=value
#         Set the value of the string variable in importpath named name to value.
#         This is only effective if the variable is declared in the source code either uninitialized
#         or initialized to a constant string expression.
export ldflags := -s -w -X main.version=$(version) -X main.branch=$(branch) -X main.commit=$(shasum) -X main.date=$(date)

# go_version is used to build the containers.
export go_version := 1.17

# aws_region denotes the region we will be pushing container images to.
export aws_region := us-west-2

# container_registry is the location where the container images will be pushed to.
# AWS_ACCOUNT and is expected to be available in the environment.
export container_registry="${AWS_ACCOUNT}.dkr.ecr.$(aws_region).amazonaws.com"

.PHONY: all
all: lint test

# ____ lint _________________________________________________________________________

.PHONY: fmt
fmt:
	@printf "\nRunning go fmt...\n"
	@go fmt ./...

.PHONY: tidy
tidy:
	@printf "\nRunning go mod tidy...\n"
	@go mod tidy

.PHONY: lint
lint: fmt tidy
	@printf "\nChecking for changed files...\n"
	@git status --porcelain
	@printf "\n"
	@if [[ "$${CI}" == "true" ]]; then $$(exit $$(git status --porcelain | wc -l)); fi

# ____ test _________________________________________________________________________

.PHONY: unit_test test

unit_test:
	@printf "\nRunning go test...\n"
	@go test -cover -race ./...

test: unit_test

# ____ container artifacts ______________________________________________________________

.PHONY: containers push_release_containers push_dev_containers

containers: Dockerfile
	DOCKER_BUILDKIT=1 docker build \
		--ssh default \
		--build-arg app=$(app) \
		--build-arg user=$(app) \
		--build-arg go_version=$(go_version) \
		--build-arg version=$(version) \
		--build-arg branch=$(branch) \
		--build-arg commit=$(commit) \
		--build-arg build_date=$(build_date) \
		--rm --force-rm \
		--tag $(app):latest \
		--tag $(app):$(container_version) \
		.
	docker image ls | grep $(app)

push_dev_containers: prefix := dev-
push_dev_containers: push_release_containers
	docker tag $(app):$(container_version) $(container_registry)/arcadium/$(app):$(branch)
	docker push $(container_registry)/arcadium/$(app):$(branch)

push_release_containers:
	aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin $(container_registry)
	docker tag $(app):$(container_version) $(container_registry)/arcadium/$(app):$(prefix)$(container_version)
	docker push $(container_registry)/arcadium/$(app):$(prefix)$(container_version)

# ____ clean artifacts ______________________________________________________________

.PHONY: clean
clean:
	@printf "\nClean...\n"
	@-rm -f dist/*
