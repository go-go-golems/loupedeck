#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

gofmt -w pkg/jsmetrics/jsmetrics.go runtime/js/runtime_test.go
go test ./pkg/jsmetrics ./runtime/js
go test ./...
