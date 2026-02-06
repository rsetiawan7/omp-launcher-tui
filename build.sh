#!/bin/bash

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT="${REPO_ROOT}/omp-tui"
VERSION="1.2.0"

echo "Building omp-launcher-tui v${VERSION}..."

cd "$REPO_ROOT"

# Use CGO_ENABLED=0 for static binary (no GUI dependencies)
export CGO_ENABLED=0

case "$(uname -s)" in
  Linux)
    GOOS=linux GOARCH=amd64 go build -o "${OUTPUT}-linux-amd64" ./cmd/omp-tui
    echo "✓ Built: ${OUTPUT}-linux-amd64"
    ;;
  Darwin)
    GOOS=darwin GOARCH=amd64 go build -o "${OUTPUT}-darwin-amd64" ./cmd/omp-tui
    GOOS=darwin GOARCH=arm64 go build -o "${OUTPUT}-darwin-arm64" ./cmd/omp-tui
    echo "✓ Built: ${OUTPUT}-darwin-amd64"
    echo "✓ Built: ${OUTPUT}-darwin-arm64"
    ;;
  MINGW*|MSYS*|CYGWIN*)
    GOOS=windows GOARCH=amd64 go build -o "${OUTPUT}-windows-amd64.exe" ./cmd/omp-tui
    echo "✓ Built: ${OUTPUT}-windows-amd64.exe"
    ;;
  *)
    echo "Unknown OS: $(uname -s)"
    exit 1
    ;;
esac

echo "Build complete!"
