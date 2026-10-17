#!/bin/bash

set -x
set -e

mkdir -p ./.vscode
: <<'EOF'
if [ ! -f "./.vscode/launch.json" ]; then
    cp -rf /data/home/project/go-tool/.vscode/launch.json ./.vscode
fi
if [ ! -f "./.vscode/tasks.json" ]; then
    cp -rf /data/home/project/go-tool/.vscode/tasks.json ./.vscode
fi
EOF
cp -rf /data/home/project/go-tool/.vscode/settings.json ./.vscode

cp -rf /data/home/project/go-tool/.claude ./
