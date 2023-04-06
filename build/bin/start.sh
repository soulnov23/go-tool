#!/bin/bash

ulimit -c unlimited
export GOTRACEBACK=crash

source ./VERSION
export GO_TOOL_VERSION=${GO_TOOL_VERSION}

./go-tool -conf ../conf/go_tool.yaml