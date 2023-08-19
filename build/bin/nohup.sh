#!/bin/bash

source init.sh

killall -SIGINT go-tool
nohup ./go-tool -conf ../conf/go_tool.yaml >/dev/null 2>&1 &
ps -ef | grep go-tool | grep -v grep
