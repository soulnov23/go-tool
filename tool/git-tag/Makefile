include ../../Inc.mk

SRC := ./
BIN := ${GOPATH}/bin/git-tag

all:
	${CGO} go build ${PRINT} -o ${BIN} ${SRC}

debug:
	${CGO} go build ${PRINT} -gcflags "$(DEBUG_GCFLAGS)" -o ${BIN} ${SRC}

release:
	${CGO} go build ${PRINT} -ldflags "$(RELEASE_LDFLAGS)" -o ${BIN} ${SRC}

.PHONY: all debug release

.DEFAULT_GOAL := all
