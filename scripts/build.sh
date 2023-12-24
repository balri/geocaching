#!/bin/bash
set -eufo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "$0")" && pwd)"
PROJ_DIR="$(readlink -f "$SCRIPT_DIR/..")"
PROJECT_NAME="${PROJECT_NAME:-geocaching}"
IMAGE_TAG="$(git rev-parse HEAD)"

cd "$PROJ_DIR"

docker_build() {
	local dockerfile="docker/$1"
	local image="$2"
	local target="$3"

	echo "Building $dockerfile"
	DOCKER_BUILDKIT=0 \
		docker build \
		--pull \
		--force-rm \
		--target="$target" \
		--build-arg "GOPROXY" \
		--build-arg "GOPRIVATE" \
		--build-arg "IMAGE_TAG=$IMAGE_TAG" \
		-t "$image:$IMAGE_TAG" \
		-f "$dockerfile" \
		.
}

docker_build 'Dockerfile' "$PROJECT_NAME" 'final'
