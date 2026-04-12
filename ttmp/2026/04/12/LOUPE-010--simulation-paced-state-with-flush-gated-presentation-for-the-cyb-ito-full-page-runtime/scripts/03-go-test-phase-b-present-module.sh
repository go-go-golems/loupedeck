#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

gofmt -w runtime/js/env/env.go runtime/js/runtime.go runtime/js/runtime_test.go runtime/js/module_present/module.go
go test ./runtime/js/...
go test ./...
