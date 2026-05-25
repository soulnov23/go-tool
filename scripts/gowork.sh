#!/bin/bash

set -x
set -e

main() {
    #go mod init $(basename $(pwd)) || true
    go work init || true
    go work use -r ./
}

main "$@"
