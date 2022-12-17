#!/bin/bash

ulimit -c unlimited
export GOTRACEBACK=crash

./go-tool -conf ../conf/go_tool.yaml