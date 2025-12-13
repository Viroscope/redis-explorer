#!/bin/bash

# Build script for Redis Explorer
# This script builds release binaries for the current platform

set -e

VERSION=${1:-v1.0.0}
OUTPUT_DIR=${2:-./releases}

echo "Building Redis Explorer ${VERSION}"
echo "Output directory: ${OUTPUT_DIR}"

mkdir -p "${OUTPUT_DIR}"

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

OUTPUT_NAME="redis-explorer-${VERSION}-${GOOS}-${GOARCH}"

echo "Building for ${GOOS}/${GOARCH}..."

if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
fi

# Build with optimizations
go build -v -ldflags="-s -w" -o "${OUTPUT_DIR}/${OUTPUT_NAME}" .

echo "Built: ${OUTPUT_DIR}/${OUTPUT_NAME}"

# Create archive
cd "${OUTPUT_DIR}"
if [ "$GOOS" = "windows" ]; then
    zip "${OUTPUT_NAME%.exe}.zip" "${OUTPUT_NAME}" ../README.md ../icon.png
    echo "Created: ${OUTPUT_NAME%.exe}.zip"
else
    tar -czf "${OUTPUT_NAME}.tar.gz" "${OUTPUT_NAME}" -C .. README.md icon.png
    echo "Created: ${OUTPUT_NAME}.tar.gz"
fi

cd ..

echo "Build complete!"
ls -lh "${OUTPUT_DIR}"
