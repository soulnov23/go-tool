#!/bin/bash

set -x
set -e

# (644)rw-r--r--
find . -type f -exec chmod 644 {} +
# (755)rwxr-xr-x
find . -type d -exec chmod 755 {} +
find . -type f -path "*/build/bin/*" ! -path "*/build/bin/*/*" -print -exec chmod +x {} +
find . -type f -name "*.sh" -print -exec chmod +x {} +
