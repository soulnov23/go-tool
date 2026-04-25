#!/bin/bash

set -x
set -e

main() {
    go mod init $(basename $(pwd)) || true
}

main "$@"
