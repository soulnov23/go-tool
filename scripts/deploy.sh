#!/bin/bash

: <<!
LOG_FILE="deploy.log"
>"${LOG_FILE}"
exec &>>${LOG_FILE}
!

set -x
set -e

WORKSPACE=$(pwd | awk -F'/go-tool' '{print $1}')

#./deploy.sh golang 1.24.6
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

#./deploy.sh protoc v29.3
function protoc() {
    mkdir -p tmp
    cd tmp
    wget https://github.com/protocolbuffers/protobuf/releases/download/$1/protoc-${1#v}-linux-x86_64.zip
    unzip protoc-${1#v}-linux-x86_64.zip
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
    go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
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

#./deploy.sh nvm v0.40.3
function nvm() {
    # Download and install nvm:
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/$1/install.sh | bash
    # in lieu of restarting the shell
    \. "$HOME/.nvm/nvm.sh"
    # Download and install Node.js:
    nvm install 24
    nvm alias default 24.11.0
    # Verify the Node.js version:
    node -v # Should print "v24.11.0".
    # Download and install pnpm:
    corepack enable pnpm
    # Verify pnpm version:
    pnpm -v
}

#./deploy.sh git v2.51.0
function git() {
    mkdir -p tmp
    cd tmp
    FILE=$1.tar.gz
    wget https://github.com/git/git/archive/refs/tags/${FILE}
    tar -zvxf ${FILE}
    cd git-${1#v}
    make -j32 prefix=/usr all
    make -j32 prefix=/usr install
    cd ../../
    rm -rf tmp
}

#./deploy.sh chromium
function chromium() {
    mkdir -p tmp
    cd tmp
    wget https://download-chromium.appspot.com/dl/Linux_x64?type=snapshots -O chromium.zip
    unzip chromium.zip
    # chromedp.ExecPath("/usr/local/bin/chrome-linux/chrome")
    rm -rf /usr/local/bin/chrome-linux || true
    mv -f chrome-linux /usr/local/bin/
    # yum install -y alsa-lib atk at-spi2-atk mesa-libgbm
    cd ..
    rm -rf tmp
}

#./deploy.sh python
function python() {
    yum install -y python3.12 python3.12-pip
    pip3.12 install --upgrade pip
}

#./deploy.sh venv /data/home/venv
function venv() {
    python3.12 -m venv $1
    source $1/bin/activate
    python3.12 install -r requirements.txt
}

#./deploy.sh markitdown v0.1.3
function markitdown() {
    mkdir -p tmp
    cd tmp
    FILE=$1.tar.gz
    wget https://github.com/microsoft/markitdown/archive/refs/tags/${FILE}
    tar -zvxf ${FILE}
    cd markitdown-${1#v}
    pip3.12 install 'packages/markitdown[docx,xls,xlsx,pptx,pdf]'
    cd ../../
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
    nvm)
        nvm $2
        ;;
    git)
        git $2
        ;;
    chromium)
        chromium
        ;;
    python)
        python
        ;;
    venv)
        venv $2
        ;;
    markitdown)
        markitdown $2
        ;;
    *)
        echo "error:argument is invalid"
        ;;
    esac
}

main "$@"
