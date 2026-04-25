#!/bin/bash

set -x
set -e

main() {
    mkdir -p ./.vscode
    if [ ! -f "./.vscode/c_cpp_properties.json" ]; then
        cp -rf /data/home/project/go-tool/.vscode/c_cpp_properties.json ./.vscode
    fi
    if [ ! -f "./.vscode/launch.json" ]; then
        cp -rf /data/home/project/go-tool/.vscode/launch.json ./.vscode
    fi
    if [ ! -f "./.vscode/tasks.json" ]; then
        cp -rf /data/home/project/go-tool/.vscode/tasks.json ./.vscode
    fi
    cp -rf /data/home/project/go-tool/.vscode/settings.json ./.vscode
    cp -rf /data/home/project/go-tool/.claude ./
    goinit ./
    shellfmt ./
    chmodinit ./
}

main "$@"
