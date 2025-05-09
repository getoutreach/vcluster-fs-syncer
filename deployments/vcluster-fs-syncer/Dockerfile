# syntax=docker/dockerfile:1.0-experimental
FROM 182192988802.dkr.ecr.us-east-2.amazonaws.com/outreach-docker/golang:1.23.4 as builder
ARG VERSION
ENV GOCACHE "/go-build-cache"
ENV GOPRIVATE github.com/getoutreach/*
ENV CGO_ENABLED 0
WORKDIR /src

# Copy our source code into the container for building
COPY . .

## <<Stencil::Block(beforeBuild)>>

## <</Stencil::Block>>

# Cache dependencies across builds
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg go mod download

# Build our application, caching the go build cache, but also using
# the dependency cache from earlier.
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg --mount=type=cache,target=/go-build-cache \
  mkdir -p bin; \
  go build -o /src/bin/ -ldflags "-X github.com/getoutreach/gobox/pkg/app.Version=$VERSION" -v ./cmd/...

FROM 182192988802.dkr.ecr.us-east-2.amazonaws.com/outreach-docker/alpine:3.18
ENTRYPOINT ["/usr/local/bin/vcluster-fs-syncer"]

LABEL "io.outreach.reporting_team"="fnd-dt"
LABEL "io.outreach.repo"="vcluster-fs-syncer"

# Add timezone information.
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO=/zoneinfo.zip

# Install certificates for RDS connectivity.
RUN wget --output-document /usr/local/share/ca-certificates/global-bundle.pem \
  "https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem" \
  && update-ca-certificates

## <<Stencil::Block(afterBuild)>>

## <</Stencil::Block>>

COPY --from=builder /src/bin/vcluster-fs-syncer /usr/local/bin/vcluster-fs-syncer
USER systemuser
