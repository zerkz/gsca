#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Building gsca..."
go build -o gsca .

echo "Done: ./gsca"
