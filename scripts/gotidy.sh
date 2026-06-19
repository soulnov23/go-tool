#!/bin/bash

set -x
set -e

go mod tidy && go get -u ./... && go mod tidy
