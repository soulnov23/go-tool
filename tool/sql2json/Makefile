include ../../Inc.mk

SRC := ./
BIN := ${GOPATH}/bin/sql2json

all:
	${CGO} go build ${PRINT} -ldflags "$(VERSION_LDFLAGS)" -o ${BIN} ${SRC}

debug:
	${CGO} go build ${PRINT} -ldflags "$(VERSION_LDFLAGS)" -gcflags "$(DEBUG_GCFLAGS)" -o ${BIN} ${SRC}

release:
	${CGO} go build ${PRINT} -ldflags "$(VERSION_LDFLAGS) $(RELEASE_LDFLAGS)" -o ${BIN} ${SRC}

.PHONY: all debug release

.DEFAULT_GOAL := all