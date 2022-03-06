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

go_version := 1.18rc1
migrate_version := v4.15.1
mkcert_version := v1.4.3

container_registry := ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# ____ all __________________________________________________________________

.PHONY: all

all: containers tools
	docker image ls

# ____ application ___________________________________________________________

.PHONY: containers arcade_container assets_container migrate_container

containers: arcade_container assets_container migrate_container

arcade_container: Dockerfile.arcade
	DOCKER_BUILDKIT=1 docker build \
		--file Dockerfile.arcade \
		--ssh default \
		--build-arg go_version=$(go_version) \
		--build-arg version=$(version) \
		--build-arg branch=$(branch) \
		--build-arg commit=$(commit) \
		--build-arg build_date=$(date) \
		--rm --force-rm \
		--tag arcade:latest \
		--tag arcade:$(version) \
		..

assets_container: Dockerfile.assets
	DOCKER_BUILDKIT=1 docker build \
		--file Dockerfile.assets \
		--ssh default \
		--build-arg go_version=$(go_version) \
		--build-arg version=$(version) \
		--build-arg branch=$(branch) \
		--build-arg commit=$(commit) \
		--build-arg build_date=$(date) \
		--rm --force-rm \
		--tag assets:latest \
		--tag assets:$(version) \
		..

migrate_container: Dockerfile.migrate
	docker build \
		--file Dockerfile.migrate \
		--build-arg migrate_version=$(migrate_version) \
		--rm --force-rm \
		--tag migrate:latest \
		--tag migrate:$(version) \
		..

# ____ push _________________________________________________________________

.PHONY: push_containers push_login push_arcade push_assets push_migrate

push_containers: push_login push_arcade push_assets push_migrate

push_login:
	aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin $(container_registry)

push_arcade:
	docker tag arcade:$(version) $(container_registry)/arcadium/arcade:$(version)
	docker push $(container_registry)/arcadium/arcade:$(version)

push_assets:
	docker tag assets:$(version) $(container_registry)/arcadium/assets:$(version)
	docker push $(container_registry)/arcadium/assets:$(version)

push_migrate:
	docker tag migrate:$(version) $(container_registry)/arcadium/migrate:$(prefix)$(version)
	docker push $(container_registry)/arcadium/migrate:$(prefix)$(version)

# ____ tools _________________________________________________________________

.PHONY: tools mkcert curl

tools: mkcert curl

mkcert: Dockerfile.mkcert
	docker build \
		--file Dockerfile.mkcert \
		--build-arg go_version=$(go_version) \
		--build-arg mkcert_version=$(mkcert_version) \
		--rm --force-rm \
		--tag mkcert:latest \
		.

curl: Dockerfile.curl
	docker build \
		--file Dockerfile.curl \
		--rm --force-rm \
		--tag curl:latest \
		.
