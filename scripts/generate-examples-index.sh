#!/usr/bin/env bash
# Regenerate examples/index.html from internal/gallery/catalog.html and examples/*.json.
set -euo pipefail
cd "$(dirname "$0")/.."
exec go run ./cmd/gobl.html generate-index
