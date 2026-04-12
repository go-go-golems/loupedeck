#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

gofmt -w runtime/metrics/metrics.go runtime/metrics/metrics_test.go
go test ./runtime/metrics/...
go test ./...
