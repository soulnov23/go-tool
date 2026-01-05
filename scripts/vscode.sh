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
    chmod -R 644 ./
    find . -type d -name build | while read -r dir; do
        chmod +x "${dir}/bin"/*
    done
    find . -type f -name "*.sh" | while read -r file; do
        chmod +x "${file}"
    done
}

main "$@"