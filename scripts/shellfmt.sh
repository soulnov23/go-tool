#!/bin/bash

set -x
set -e

main() {
    if [[ $# -ne 1 ]]; then
        echo "Usage: $0 <path>"
        exit 1
    fi
    # --language-dialect
    # --indent
    # --case-indent switch cases will be indented
    # --write write result to file instead of stdout
    shfmt -ln bash -i 4 -ci -w $1
}

main "$@"
