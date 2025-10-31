include ../../Inc.mk

SRC	:= ./
BIN := ../../build/bin/git-tag

all:
	${CGO} go build ${PRINT} ${VERSION} -o ${BIN} ${SRC}

debug:
	${CGO} go build ${PRINT} ${VERSION} ${DEBUG} -o ${BIN} ${SRC}

release:
	${CGO} go build ${PRINT} ${VERSION} ${RELEASE} -o ${BIN} ${SRC}

.PHONY: all debug release

.DEFAULT_GOAL: all