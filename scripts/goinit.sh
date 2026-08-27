#!/bin/bash

set -x
set -e

main() {
    rm -rf go.work go.work.sum
    go mod init $(basename $(pwd)) || true
}

main "$@"
