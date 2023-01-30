#!/bin/bash

ulimit -c unlimited
export GOTRACEBACK=crash

source ./VERSION
export SERVER_VERSION=${SERVER_VERSION}

./go-tool -conf ../conf/go_tool.yaml