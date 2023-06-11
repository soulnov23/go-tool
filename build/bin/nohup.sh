#!/bin/bash

ulimit -c unlimited
export GOTRACEBACK=crash

pid=$(ps -ef | grep go-tool | grep -v grep | awk '{print $2}')
if [ -n "${pid}" ]; then
    kill -SIGINT ${pid}
fi
nohup ./go-tool -conf ../conf/go_tool.yaml >/dev/null 2>&1 &
ps -ef | grep go-tool | grep -v grep
