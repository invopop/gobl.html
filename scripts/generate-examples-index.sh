#!/usr/bin/env bash
# Regenerate examples/index.html (same as: go test -run TestGOBLRenderExamples -update, which also refreshes examples/out).
set -euo pipefail
cd "$(dirname "$0")/.."
exec go run ./cmd/gobl.html generate-index
