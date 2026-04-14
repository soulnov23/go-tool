#!/bin/bash

set -x
set -e

main() {
    go mod tidy && go get -u ./... && go mod tidy
}

main "$@"
