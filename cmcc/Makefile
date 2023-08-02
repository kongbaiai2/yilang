TARGETS := cmcc
SOURCES := $(wildcard *.go)

LINUX_64   := ../bin/${TARGETS}_linux64
LINUX_arm  := ../bin/${TARGETS}_linuxarm
WINDOWS_64 := ../bin/${TARGETS}_win64.exe
DARWIN_64  := ../bin/${TARGETS}_darwin64

# GO_PATH := $(shell pwd)/../vendor/
GO111MODULE=auto
BUILD_TIME=`date +%FT%T`
GIT_REVISIONS=`git rev-parse --short HEAD`
LDFLAGS=-ldflags "-X main.buildTimestamp=${BUILD_TIME} -X main.buildGitRevision=${GIT_REVISIONS}"

.PHONY: all run clean help

## linux: 编译打包linux
.PHONY: linux
linux:
	GO111MODULE=$(GO111MODULE) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(RACE) ${LDFLAGS} -o ${LINUX_64} ./main.go

## mac: 编译打包mac
.PHONY: mac
mac:
	GO111MODULE=$(GO111MODULE) CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(RACE) ${LDFLAGS} -o ${DARWIN_64} ./main.go

## linux: 编译ARM64 打包linux
.PHONY: linuxarm
linuxarm:
	GO111MODULE=$(GO111MODULE) CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(RACE) ${LDFLAGS} -o ${LINUX_arm} ./main.go
 
## win: 编译打包win
.PHONY: win
win:
	GO111MODULE=$(GO111MODULE) CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(RACE) ${LDFLAGS} -o ${WINDOWS_64} ./main.go


build:
	@go build -o ${TARGETS}
 
## 编译win，linux，mac平台
.PHONY: all
all:win linux mac linuxarm
 
run:
	@go run ./
 
.PHONY: tidy
tidy:
	@go mod tidy
 
## test: Run unit test.
.PHONY: test
test:
	@$(MAKE) go.test
 
## 清理二进制文件
clean:
	@if [ -f ${LINUX_64} ] ; then rm ${LINUX_64}; fi
	@if [ -f ${LINUX_arm} ] ; then rm ${LINUX_arm}; fi
	@if [ -f ${WINDOWS_64} ] ; then rm ${WINDOWS_64}; fi
	@if [ -f ${DARWIN_64} ] ; then rm ${DARWIN_64}; fi
 
help:
	@echo "make 			- 格式化 Go 代码, 并编译生成二进制文件"
	@echo "make mac 		- 编译 Go 代码, 生成mac的二进制文件"
	@echo "make linux 		- 编译 Go 代码, 生成linux二进制文件"
	@echo "make linuxarm		- 编译 Go 代码, 生成linux arm二进制文件"
	@echo "make win 		- 编译 Go 代码, 生成windows二进制文件"
	@echo "make tidy 		- 执行go mod tidy"
	@echo "make run 		- 直接运行 Go 代码"
	@echo "make clean 		- 移除编译的二进制文件"
	@echo "make all 		- 编译多平台的二进制文件"
