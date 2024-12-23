#!/bin/bash

: <<!
LOG_FILE="deploy.log"
>"${LOG_FILE}"
exec &>>${LOG_FILE}
!

set -x
set -e

WORKSPACE=$(pwd | awk -F'/go-tool' '{print $1}')

#./deploy.sh golang 1.23.1
function golang() {
    mkdir -p tmp
    cd tmp
    FILE=go$1.tar.gz
    wget https://github.com/golang/go/archive/refs/tags/${FILE}
    tar -zvxf ${FILE}
    cd go-go$1/src
    ./make.bash
    cd -
    mv ${FILE} go-go$1/
    GOROOT=$(go env GOROOT)
    rm -rf ${GOROOT}
    cp -rf go-go$1 ${GOROOT}
    cd ..
    rm -rf tmp
}

#./deploy.sh protoc 24.4
function protoc() {
    mkdir -p tmp
    cd tmp
    wget https://github.com/protocolbuffers/protobuf/releases/download/v$1/protoc-$1-linux-x86_64.zip
    unzip protoc-$1-linux-x86_64.zip
    for FILE in ./bin/*; do
        cp -rf ${FILE} /usr/local/bin
    done
    for FILE in ./include/*; do
        cp -rf ${FILE} /usr/local/include
    done
    cd ..
    rm -rf tmp
    go install github.com/golang/protobuf/protoc-gen-go@latest
    go install github.com/envoyproxy/protoc-gen-validate@latest
    go install github.com/golang/mock/mockgen@latest
}

#./deploy.sh kubectl v1.18.4
#kubectl版本和集群版本之间的差异必须在一个小版本号内。例如：v1.30版本的客户端能与v1.29、v1.30和v1.31版本的控制面通信。用最新兼容版的kubectl有助于避免不可预见的问题
function kubectl() {
    mkdir -p tmp
    cd tmp
    curl -LO "https://dl.k8s.io/release/$1/bin/linux/amd64/kubectl"
    curl -LO "https://dl.k8s.io/release/$1/bin/linux/amd64/kubectl.sha256"
    echo "$(cat kubectl.sha256) kubectl" | sha256sum --check
    GOPATH=$(go env GOPATH)
    cp -rf kubectl ${GOPATH}/bin
    git clone https://github.com/ahmetb/kubectx
    cp -rf ./kubectx/kubectx ${GOPATH}/bin
    cp -rf ./kubectx/kubens ${GOPATH}/bin
    go install github.com/derailed/k9s@latest
    cd ..
    rm -rf tmp
}

main() {
    case $1 in
    golang)
        golang $2
        ;;
    protoc)
        protoc $2
        ;;
    kubectl)
        kubectl $2
        ;;
    *)
        echo "error:argument is invalid"
        ;;
    esac
}

main "$@"
