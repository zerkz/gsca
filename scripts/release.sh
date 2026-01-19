#!/usr/bin/env bash
set -euo pipefail

# Release script for gsca
# Usage: ./scripts/release.sh 1.2.0
# Requires: go, golangci-lint, goreleaser

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.2.0"
    exit 1
fi

VERSION="$1"
TAG="v${VERSION}"

# Validate version format
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.2.0)"
    exit 1
fi

# Check for uncommitted changes
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "Error: You have uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Delete local tag if it exists (safe since we check remote later)
if git rev-parse "$TAG" >/dev/null 2>&1; then
    # Check if tag exists on remote
    if git ls-remote --tags origin | grep -q "refs/tags/$TAG"; then
        echo "Error: Tag $TAG already exists on remote"
        exit 1
    fi
    echo "Deleting existing local tag $TAG..."
    git tag -d "$TAG"
fi

echo "=== Releasing gsca $TAG ==="
echo

# Run tests
echo "Running tests..."
go test ./...
echo "Tests passed."
echo

# Run linter
echo "Running linter..."
golangci-lint run
echo "Linter passed."
echo

# Check goreleaser config
echo "Checking GoReleaser config..."
goreleaser check
echo

# Update PKGBUILD
echo "Updating PKGBUILD..."
sed -i "s/^pkgver=.*/pkgver=${VERSION}/" PKGBUILD
echo "  PKGBUILD updated to ${VERSION}"

# Update Flatpak manifest
echo "Updating Flatpak manifest..."
sed -i "s/tag: v.*/tag: ${TAG}/" com.github.zerkz.gsca.yaml
echo "  Flatpak manifest updated to ${TAG}"

# Update README.md install instructions
echo "Updating README.md..."
sed -i -E "s|/releases/download/v[0-9]+\.[0-9]+\.[0-9]+/|/releases/download/${TAG}/|g" README.md
sed -i -E "s|gsca-v[0-9]+\.[0-9]+\.[0-9]+\.flatpak|gsca-${TAG}.flatpak|g" README.md
sed -i -E "s|gsca_[0-9]+\.[0-9]+\.[0-9]+_|gsca_${VERSION}_|g" README.md
echo "  README.md updated to ${VERSION}"

# Show changes
echo
echo "=== Changes ==="
git diff --stat
echo
git diff

# Confirm
echo
read -p "Commit these changes and create tag $TAG? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborting. Reverting changes..."
    git checkout -- PKGBUILD com.github.zerkz.gsca.yaml README.md
    exit 1
fi

# Commit version bump (if there are changes)
git add PKGBUILD com.github.zerkz.gsca.yaml README.md
if git diff --cached --quiet; then
    echo "No version changes needed (already at ${VERSION})"
else
    echo "Committing version bump..."
    git commit -m "Bump version to ${VERSION}"
fi

# Create tag
echo "Creating tag $TAG..."
git tag -a "$TAG" -m "Release ${TAG}"

# Push
echo
read -p "Push commit and tag to origin? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipping push. You can push manually with:"
    echo "  git push origin main"
    echo "  git push origin $TAG"
    exit 0
fi

echo "Pushing..."
git push origin main --follow-tags

echo
echo "=== Release $TAG complete ==="
echo "GitHub Actions will now build and publish the release."
echo "Monitor progress at: https://github.com/zerkz/gsca/actions"
