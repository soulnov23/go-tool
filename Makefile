SRC	 := ./cmd
BIN  := ./build/bin/go-tool

#使用go tool compile查看-gcflags传递给编译器的参数
DEBUG 	:= -gcflags "all=-N -l"
#使用go tool link查看-ldflags传递给链接器的参数
RELEASE := -ldflags "-w -s"

VERSION := -ldflags "-X 'main.goVersion=$(shell go version)' \
					 -X 'main.gitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
					 -X 'main.gitCommitID=$(shell git rev-parse HEAD)' \
					 -X 'main.gitCommitTime=$(shell git log --pretty=format:"%ci" | head -1)' \
					 -X 'main.gitCommitAuthor=$(shell git log --pretty=format:"%cn" | head -1)'" \

CGO := CGO_ENABLED=0

all:
	${CGO} go build ${VERSION} ${DEBUG} -o ${BIN} ${SRC}

clean:
	rm -rf ${BIN}

.PHONY: all clean

.DEFAULT_GOAL: all