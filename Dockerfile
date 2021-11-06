# Build
ARG go_version
FROM golang:${go_version}-alpine as builder
LABEL stage=intermediate

ARG app
ENV GOPRIVATE arcadium.dev

WORKDIR $GOPATH/src/arcadium.dev/${app}

# Create a layer for accessing a private github repo via ssh.
RUN apk update && apk add --no-cache openssh-client git && \
    git config --global url."git@github.com:".insteadOf "https://github.com/" && \
    mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Create a separate layer for the go dependencies to facilitate faster builds.
COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download && go mod verify

# Create a layer for the build.
COPY . .

ARG user
ARG version
ARG branch
ARG commit
ARG build_date

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w \
      -X 'main.version=${version}' \
      -X 'main.branch=${branch}' \
      -X 'main.commit=${commit}' \
      -X 'main.date=${build_date}'" \
      -o bin/${app} ./cmd/${app}


# Assemble the debug image
FROM alpine as debug

ARG app
ARG user

RUN apk update && \
    apk add --no-cache ca-certificates tzdata bash && update-ca-certificates && \
    adduser \
      --disabled-password \
      --gecos "" \
      --home "/nonexistent" \
      --shell "/sbin/nologin" \
      --no-create-home \
      --uid "10001" \
      ${user}

COPY --from=builder /go/src/arcadium.dev/${app}/bin/${app} /usr/local/bin/${app}

USER ${user}:${user}
EXPOSE 8443
ENTRYPOINT ["/usr/local/bin/${app}"]


# Assemble the image
FROM scratch

ARG app
ARG user

COPY --from=debug /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=debug /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=debug /etc/passwd /etc/passwd
COPY --from=debug /etc/group /etc/group

COPY --from=debug /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=debug /usr/local/bin/${app} /usr/local/bin/${app}

USER ${user}:${user}
EXPOSE 8443
ENTRYPOINT ["/usr/local/bin/${app}"]
