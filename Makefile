SRC	:= ./cmd
BIN := ./build/bin/go-tool

#-v编译包时打印包的名称
#-x打印命令
PRINT := -v -x

#使用go tool compile查看-gcflags传递给编译器的参数
#-N禁用优化
#-l禁用内联
DEBUG := -gcflags "all=-N -l"

#使用go tool link查看-ldflags传递给链接器的参数
#-w禁用DWARF生成
#-s禁用符号表
RELEASE := -ldflags "-w -s"

#-m打印优化策略，编译器优化技术确定变量是否需要在堆上分配内存
ESCAPE := -gcflags "-m"

VERSION := -ldflags "-X 'main.goVersion=$(shell go version)' \
					 -X 'main.gitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
					 -X 'main.gitCommitID=$(shell git rev-parse HEAD)' \
					 -X 'main.gitCommitTime=$(shell git log --pretty=format:"%ci" | head -1)' \
					 -X 'main.gitCommitAuthor=$(shell git log --pretty=format:"%cn" | head -1)'"

CGO := CGO_ENABLED=0

all:
	${CGO} go build ${PRINT} ${VERSION} -o ${BIN} ${SRC}

debug:
	${CGO} go build ${PRINT} ${VERSION} ${DEBUG} -o ${BIN} ${SRC}

release:
	${CGO} go build ${PRINT} ${VERSION} ${RELEASE} -o ${BIN} ${SRC}

escape:
	go build ${ESCAPE} -o temp ${SRC}
	rm -rf temp

test:
	go test -v -count 1 -race -timeout 1s ./...

#go env GOCACHE
#go env GOMODCACHE
clean:
    #go clean ${PRINT} -i -cache -testcache -modcache -fuzzcache
	rm -rf ${BIN}

.PHONY: all debug release escape test clean

.DEFAULT_GOAL: all