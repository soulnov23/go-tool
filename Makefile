include Inc.mk

SRC := ./
BIN := ./build/bin/go-tool

all:
	${CGO} go build ${PRINT} -ldflags "$(VERSION_LDFLAGS)" -o ${BIN} ${SRC}

debug:
	${CGO} go build ${PRINT} -ldflags "$(VERSION_LDFLAGS)" -gcflags "$(DEBUG_GCFLAGS)" -o ${BIN} ${SRC}

release:
	${CGO} go build ${PRINT} -ldflags "$(VERSION_LDFLAGS) $(RELEASE_LDFLAGS)" -o ${BIN} ${SRC}

errors:
	protoc --proto_path ./pkg/errors --go_out paths=source_relative:./pkg/errors --validate_out lang=go,paths=source_relative:./pkg/errors errors.proto

example:
	go run ./pkg/errors/generator -source ./pkg/errors/example/common.yaml,./pkg/errors/example/errors.yaml -destination ./pkg/errors/example/errors.go -package errors

escape:
	go build ${ESCAPE} -o /dev/null ${SRC}

test:
	go test -v -count 1 -race -timeout 1s ./...

#go env GOCACHE
#go env GOMODCACHE
clean:
	#go clean ${PRINT} -i -cache -testcache -modcache -fuzzcache
	rm -rf ${BIN}

#(644)rw-r--r--
#(755)rwxr-xr-x
chmod:
	chmod -R 644 ./
	chmod +x ./build/bin/* ./scripts/*

docker:
	docker build -t go-tool:latest .
	docker run --rm -it --entrypoint /bin/bash go-tool:latest

.PHONY: all debug release errors example escape test clean chmod docker

.DEFAULT_GOAL := all