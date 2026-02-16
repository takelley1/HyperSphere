#!/usr/bin/env bash
# Path: scripts/test.sh
# Description: Run unit tests with an enforced 100 percent coverage gate for internal packages.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

GOCACHE="${ROOT_DIR}/.gocache" go test ./cmd/...
GOCACHE="${ROOT_DIR}/.gocache" go test ./internal/... -coverprofile=coverage.out
TOTAL_COVERAGE="$(GOCACHE="${ROOT_DIR}/.gocache" go tool cover -func=coverage.out | awk '/^total:/ {print $3}')"
if [[ "${TOTAL_COVERAGE}" != "100.0%" ]]; then
  printf 'coverage check failed: expected 100.0%%, got %s\n' "${TOTAL_COVERAGE}"
  exit 1
fi

printf 'tests passed with total coverage %s\n' "${TOTAL_COVERAGE}"
