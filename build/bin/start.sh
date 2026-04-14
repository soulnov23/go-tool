#!/bin/bash

source init.sh

killall -SIGINT go-tool
./go-tool -conf ../conf/go_tool.yaml
