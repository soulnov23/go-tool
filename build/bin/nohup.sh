#!/bin/bash

ulimit -c unlimited
export GOTRACEBACK=crash

killall -9 go-tool
nohup ./go-tool -conf ../conf/go_tool.yaml >/dev/null 2>&1 &
ps -ef | grep go-tool | grep -v grep | grep -v vi | grep -v tail | grep -v kill
