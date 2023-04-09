SRC	 := ./cmd
BIN  := ./build/bin/go-tool

DEBUG 	:= -gcflags "all=-N -l"
RELEASE := -ldflags "-w -s"

VERSION := -ldflags "-X 'main.goVersion=$(shell go version)' \
					 -X 'main.gitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
					 -X 'main.gitCommitID=$(shell git rev-parse HEAD)' \
					 -X 'main.gitCommitTime=$(shell git show --pretty=format:"%ci %cr" | head -1)'"

CGO := CGO_ENABLED=0

all:
	${CGO} go build ${VERSION} ${DEBUG} -o ${BIN} ${SRC}

clean:
	rm -rf ${BIN}

.PHONY: all clean

.DEFAULT_GOAL: all