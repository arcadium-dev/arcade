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

go_version := 1.22
migrate_version := v4.17.1
mkcert_version := v1.4.4

# ____ all __________________________________________________________________

.PHONY: all

all: images tools
	docker image ls

# ____ application ___________________________________________________________

.PHONY: images assets migrate

images: assets migrate

assets: Dockerfile.assets
	DOCKER_BUILDKIT=1 docker build \
		--file Dockerfile.assets \
		--ssh default \
		--build-arg go_version=$(go_version) \
		--build-arg arc=$(arch) \
		--build-arg version=$(version) \
		--build-arg branch=$(branch) \
		--build-arg commit=$(commit) \
		--build-arg build_date=$(date) \
		--build-arg buildargs=$(buildargs) \
		--rm --force-rm \
		--tag assets:latest \
		..

migrate: Dockerfile.migrate
	DOCKER_BUILDKIT=1 docker build \
    --file Dockerfile.migrate \
    --build-arg migrate_version=$(migrate_version) \
    --rm --force-rm \
    --tag migrate:latest \
    ..

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
