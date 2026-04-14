#!/bin/bash

set -x
set -e

main() {
    shfmt -ln bash -i 4 -ci -kp -w $@
}

main "$@"
