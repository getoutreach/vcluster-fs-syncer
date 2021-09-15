# syntax=docker/dockerfile:1.0-experimental
FROM gcr.io/outreach-docker/golang:1.15.2 as builder
ARG VERSION
ENV GOCACHE "/go-build-cache"
ENV GOPRIVATE github.com/getoutreach/*
ENV CGO_ENABLED 0
WORKDIR /src

# Copy our source code into the container for building
COPY . .

# Cache dependencies across builds
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg make dep

# Build our application, caching the go build cache, but also using
# the dependency cache from earlier.
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg --mount=type=cache,target=/go-build-cache \
    mkdir -p bin; \
    go build -o /src/bin/ -ldflags "-X github.com/getoutreach/gobox/pkg/app.Version=$VERSION" -v ./cmd/...


FROM gcr.io/outreach-docker/alpine:3.12
ENTRYPOINT ["/usr/local/bin/vcluster-fs-syncer"]

LABEL "io.outreach.reporting_team"="cia-dev-tooling"
LABEL "io.outreach.repo"="vcluster-fs-syncer"

# Add timezone information.
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO=/zoneinfo.zip

###Block(afterBuild)
###EndBlock(afterBuild)

COPY --from=builder /src/bin/vcluster-fs-syncer /usr/local/bin/vcluster-fs-syncer

USER systemuser
