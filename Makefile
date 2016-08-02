SHELL=/bin/bash -o pipefail

GO_FILES = $(shell find . -type f -name '*.go')
ARCH := amd64
BUILD_DIR := build
PROJECT := github.com/bouk/extractdata
CMD := extractdata

SERVER_BINARY := ${BUILD_DIR}/linux_${ARCH}/${CMD}
DEVELOP_BINARY := ${BUILD_DIR}/darwin_${ARCH}/${CMD}

build: ${SERVER_BINARY} ${DEVELOP_BINARY}

${BUILD_DIR}/%_${ARCH}/${CMD}: ${GO_FILES}
	@mkdir -p "$(@D)"
	GOARCH=${ARCH} GOOS=$* go build -ldflags="-w -s" -i -o $@ ${PROJECT}

template/ego.go: template/*.ego
	go generate ${PROJECT}/template
	go fmt template/ego.go

develop: ${DEVELOP_BINARY}
	${DEVELOP_BINARY}

.PHONY: develop
-include deploy.mk
