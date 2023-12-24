FROM golang:1.21.2-alpine3.18 AS build

ARG GOPROXY
ARG GOPRIVATE
ARG IMAGE_TAG

RUN apk add --no-cache \
	bash=~5 \
	&& rm -rf /var/cache/apk

# Dependencies
SHELL ["/bin/bash", "-eo", "pipefail", "-c"]
WORKDIR /build/
COPY go.mod go.sum ./
RUN go mod download

# The rest
COPY . .

ENV IMAGE_TAG "$IMAGE_TAG"

# Build & Test
RUN echo "Testing and compiling project" \
	&& golangci-lint run ./... \
	&& echo 'cover' \
	&& richgo test -cover -v ./... \
	&& echo 'bench' \
	&& richgo test -bench=. ./... \
	&& echo 'check vulnerability' \
	&& govulncheck ./... \
	&& echo 'build geocaching' \
	&& go build \
	-ldflags "-w -s -X main.commit=$IMAGE_TAG" \
	-o geocaching \
	-v ./cmd

# Final image
FROM scratch AS final
LABEL org.label-schema.vcs-ref=$IMAGE_TAG
COPY --from=build /build/geocaching /geocaching
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 80
ENTRYPOINT ["/geocaching"]
