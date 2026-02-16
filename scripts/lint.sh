#!/usr/bin/env bash
# Path: scripts/lint.sh
# Description: Run formatting and static checks for the HyperSphere Go codebase.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

GO_FILES="$(find cmd internal -type f -name '*.go' | sort)"
UNFORMATTED="$(gofmt -l ${GO_FILES})"
if [[ -n "${UNFORMATTED}" ]]; then
  printf 'The following files need gofmt:\n%s\n' "${UNFORMATTED}"
  exit 1
fi

GOCACHE="${ROOT_DIR}/.gocache" go vet ./...
printf 'lint checks passed\n'
