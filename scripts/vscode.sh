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
        bin_dir="${dir}/bin"
        if [ -d "$bin_dir" ]; then
            find "$bin_dir" -maxdepth 1 -type f -exec chmod +x {} \;
        fi
    done
    find . -type f -name "*.sh" | while read -r file; do
        chmod +x "${file}"
    done
}

main "$@"