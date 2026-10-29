#!/usr/bin/env bash
# Builds the files published to https://assets.invopop.com/gobl.html/<version>/
# into the given directory: the WebAssembly renderer, the wasm_exec.js from the
# Go toolchain that built it, and the stylesheets and scripts for hosts that
# render with embed_assets off.
#
#   scripts/cdn-dist.sh <dir> [version]
set -euo pipefail

out=${1:?usage: scripts/cdn-dist.sh <dir> [version]}
version=${2:-dev}
cd "$(dirname "$0")/.."
mkdir -p "$out"

GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w -X main.version=$version" \
  -o "$out/gobl.html.wasm" ./wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/"
cp assets/styles/*.css assets/scripts/*.js "$out/"

ls -l "$out"
