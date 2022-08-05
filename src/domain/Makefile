TARGETS := go_domain
SOURCES := $(wildcard *.go)

BUILD_TIME=`date +%FT%T`
GIT_REVISIONS=`git rev-parse --short HEAD`
LDFLAGS=-ldflags "-X main.buildTimestamp=${BUILD_TIME} -X main.buildGitRevison=${GIT_REVISIONS}"

GO_PATH := $(dir $(abspath $(dir $$PWD)))vendor/
GO111MODULE=auto

all: $(TARGETS)

$(TARGETS): $(SOURCES)
	GOPATH=$(GO_PATH) GOOS=$(os) GOARCH=$(arch)  go build ${LDFLAGS} -o ${TARGETS}

.PHONY: clean

clean:
	rm -rf ${TARGETS}

