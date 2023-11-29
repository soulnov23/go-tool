#!/bin/bash

: <<!
LOG_FILE="deploy.log"
>"${LOG_FILE}"
exec &>>${LOG_FILE}
!

set -x
set -e

WORKSPACE=$(pwd | awk -F'/go-tool' '{print $1}')

#./deploy.sh golang 1.21.3
function golang() {
    mkdir -p tmp
    cd tmp
    FILE=go$1.src.tar.gz
    wget https://go.dev/dl/${FILE}
    tar -zvxf ${FILE}
    cd go/src
    ./make.bash
    cd -
    mv ${FILE} go/
    GOROOT=$(go env GOROOT)
    rm -rf ${GOROOT}
    cp -rf go ${GOROOT}
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

function jaeger() {
    docker run -d --name jaeger \
        -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
        -p 5775:5775/udp \
        -p 6831:6831/udp \
        -p 6832:6832/udp \
        -p 5778:5778 \
        -p 16686:16686 \
        -p 14268:14268 \
        -p 14250:14250 \
        -p 9411:9411 \
        jaegertracing/all-in-one:latest
}

function prometheus() {
    docker run -d --name prometheus \
        -p 9090:9090 \
        -v /etc/prometheus/:/etc/prometheus/ \
        prom/prometheus
}

function grafana() {
    docker run -d --name grafana \
        -p 3000:3000 \
        -v /etc/grafana/:/etc/grafana/ \
        grafana/grafana
}

main() {
    case $1 in
    golang)
        golang $2
        ;;
    protoc)
        protoc $2
        ;;
    jaeger)
        jaeger
        ;;
    prometheus)
        prometheus
        ;;
    grafana)
        grafana
        ;;
    *)
        echo "error:argument is invalid"
        ;;
    esac
}

main "$@"
