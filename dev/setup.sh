#!/usr/bin/env bash

CURRENT_DIR="$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)"

set -ex
set -o pipefail

source "${CURRENT_DIR}/common.sh"
go install sigs.k8s.io/kind@v0.29.0

# Default image
KIND_IMAGE="quay.io/powercloud/kind-node:v1.33.1"

# Override for arm64
if [[ "$(uname -m)" == "arm64" ]]; then
  KIND_IMAGE="kindest/node:latest"
fi

# Allow passing image as argument (optional override)
KIND_IMAGE="${1:-$KIND_IMAGE}"

# Create the cluster
kind create cluster \
  --image "${KIND_IMAGE}" \
  --name "${KIND_CLUSTER_NAME}" \
  --config "${KIND_CLUSTER_CONFIG_PATH}" \
  --wait 5m