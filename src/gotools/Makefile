
TARGETS := gotools
SOURCES := $(wildcard *.go)

RELEASE := 1

BUILD_TIME=`date +%FT%T`
#GIT_REVISIONS=`git rev-parse --short HEAD`
#LDFLAGS=-ldflags "-X main.buildTimestamp=${BUILD_TIME} -X main.buildGitRevison=${GIT_REVISIONS}"

LDFLAGS=-ldflags "-X main.buildTimestamp=${BUILD_TIME} " #-X main.buildGitRevison=${GIT_REVISIONS}"

GO_PATH := $(dir $(abspath $(dir $$PWD)))vendor/
# GO_PATH := "/Users/kongbaiai2/go/pkg"
GO111MODULE=auto

all: $(TARGETS)

$(TARGETS): $(SOURCES)
	GOPATH=$(GO_PATH) GOOS=$(os) go build -i ${LDFLAGS} -o ${TARGETS}

.PHONY: clean

clean:
	rm -rf ${TARGETS}
