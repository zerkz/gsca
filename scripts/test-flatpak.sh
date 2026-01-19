#!/usr/bin/env bash
set -euo pipefail

# Test flatpak build locally using act
# Usage: sudo ./scripts/test-flatpak.sh [tag]
# Example: sudo ./scripts/test-flatpak.sh v1.0.0

TAG="${1:-v1.0.0}"

cd "$(dirname "$0")/.."

# Check for GITHUB_TOKEN
if [ -z "${GITHUB_TOKEN:-}" ]; then
    echo "GITHUB_TOKEN not set. Trying gh auth token..."
    GITHUB_TOKEN=$(gh auth token 2>/dev/null || true)
    if [ -z "$GITHUB_TOKEN" ]; then
        echo "Error: GITHUB_TOKEN required. Set it or run 'gh auth login' first."
        exit 1
    fi
fi

echo "Testing flatpak build for $TAG"
echo

sudo GITHUB_TOKEN="$GITHUB_TOKEN" act workflow_dispatch \
    -W .github/workflows/test-flatpak.yml \
    -j flatpak \
    --input tag="$TAG" \
    -s GITHUB_TOKEN="$GITHUB_TOKEN" \
    --container-options "--privileged"
