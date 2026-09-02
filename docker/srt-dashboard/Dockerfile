FROM alpine:latest AS runner-setup

RUN addgroup -S srt \
 && adduser  -S srt \
             -G srt \
             -h /opt/srt \
             -s /sbin/nologin \
             -D \
 && mkdir -p /opt/srt \
 && chown srt:srt /opt/srt

FROM --platform=$BUILDPLATFORM golang:alpine AS build-0

RUN apk add --no-cache libcap ca-certificates

# Set up working directory
WORKDIR /app

# Copy go.mod and go.sum separately so we only invalidate the downloading layers if we need to
COPY go.mod go.sum ./

# Fetch dependencies and build the binary
ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download -x

# Copy the rest of the project to ensure code changes dont trigger re-download of all deps
COPY . .

FROM --platform=$BUILDPLATFORM build-0 AS build-dashboard

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o srt-dashboard_${TARGETOS}_${TARGETARCH} -ldflags="-w -s -X main.version=${GIT_TAG}" ./cmd/srt-dashboard && \
  setcap 'cap_net_bind_service=+ep' srt-dashboard_${TARGETOS}_${TARGETARCH}

# Runner 
FROM scratch

COPY --from=runner-setup /etc/passwd /etc/passwd
COPY --from=runner-setup /etc/group /etc/group
COPY --from=build-0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENV HOME=/opt/srt

# Set up the working directory
WORKDIR /opt/srt

ARG TARGETOS
ARG TARGETARCH

# Copy the binary from the builder stage
COPY --from=build-dashboard --chmod=777 /app/srt-dashboard_${TARGETOS}_${TARGETARCH} /usr/bin/srt-dashboard

USER srt

EXPOSE 8080

ENTRYPOINT ["/usr/bin/srt-dashboard"]
