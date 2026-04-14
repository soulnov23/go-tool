#!/bin/bash

set -x
set -e

main() {
    # --language-dialect
    # --indent
    # --case-indent switch cases will be indented
    # --write write result to file instead of stdout
    shfmt -ln bash -i 4 -ci -w $@
}

main "$@"
