#!/bin/bash

set -x
set -e

#go mod init $(basename $(pwd)) || true
go work init || true
go work use -r ./
