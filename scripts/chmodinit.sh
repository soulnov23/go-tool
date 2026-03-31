#!/bin/bash

set -x
set -e

main() {
    # (644)rw-r--r--
    find . -type f -exec chmod 644 {} \;
    # (755)rwxr-xr-x
    find . -type d -exec chmod 755 {} \;
    find . -type d -path "*/build/bin" | xargs -I{} find {} -maxdepth 1 -type f -print -exec chmod +x {} \;
    find . -type f -name "*.sh" -print -exec chmod +x {} \;
}

main "$@"