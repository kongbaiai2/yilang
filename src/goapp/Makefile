# 定义默认构建目录
BLDDIR ?= _output

# 注入版本信息（支持外部传入，默认自动检测）
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "v0.0.0")
BUILD_TIME ?= $(shell date +%FT%T)
GIT_REVISION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS= -ldflags "-X main.buildTimestamp=$(BUILD_TIME) -X main.buildGitRevision=$(GIT_REVISION) -X main.tagGitVersion=$(VERSION)"

# 定义支持的平台列表（格式：OS/ARCH）
BUILD_PLATFORMS ?= \
	darwin/amd64 \
	linux/amd64

# darwin/amd64 \
# darwin/arm64 \
# linux/amd64 \
# linux/arm64 \
# windows/amd64 \
# windows/386

# 目标文件名前缀
APP_NAME = goapp

# 默认目标：构建当前平台的二进制
.DEFAULT_GOAL := build-current

# 定义伪目标
.PHONY: all clean build-current build test cover help

# 默认构建当前平台（自动检测系统）
build-current: $(BLDDIR)/$(APP_NAME)-current

# 构建所有平台
build all: $(foreach platform,$(BUILD_PLATFORMS),$(BLDDIR)/$(APP_NAME)-$(subst /,-,$(platform)))

# 清理所有构建产物
clean:
	rm -rf $(BLDDIR)/*

test:
	@echo "运行单元测试..."
	go test -v -gcflags=-l ./...

cover:
	@echo "生成测试覆盖率报告..."
	@go test -coverprofile=cover.out -gcflags=-l ./...
	@go tool cover -html=cover.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

#---------------------------- 核心规则 ----------------------------#
# 构建当前平台
$(BLDDIR)/$(APP_NAME)-current: $(wildcard apps/goapp/*.go) go.mod go.sum
	@mkdir -p $(dir $@)
	go build ${LDFLAGS} -o $@ ./apps/goapp/

# 多平台交叉编译模板定义
define build_template
$(BLDDIR)/$(APP_NAME)-$(subst /,-,$(1)): $(wildcard apps/goapp/*.go) go.mod go.sum
	@mkdir -p $$(dir $$@)
	GOOS=$(2) GOARCH=$(3) go build ${LDFLAGS} -o $$@ ./apps/goapp/
	@# 自动为 Windows 添加 .exe 后缀
	@if [ "$(2)" = "windows" ]; then mv $$@ $$@.exe; fi
endef

# 解析平台列表并生成对应规则
$(foreach platform,$(BUILD_PLATFORMS),\
	$(eval parts := $(subst /, ,$(platform))) \
	$(eval os := $(word 1,$(parts))) \
	$(eval arch := $(word 2,$(parts))) \
	$(eval $(call build_template,$(platform),$(os),$(arch))) \
)


help:
	@echo "用法:"
	@echo "  make                  构建当前平台（默认）"
	@echo "  make all              构建所有支持的平台"
	@echo "  make build BUILD_PLATFORMS=\"linux/amd64\"  构建指定平台"
	@echo "  make clean            清理构建产物"
	@echo "  make test             运行单元测试"
	@echo "  make cover            生成测试覆盖率报告"
	@echo ""
	@echo "当前支持的平台 (BUILD_PLATFORMS):"
	@echo "$(BUILD_PLATFORMS)" | tr ' ' '\n' | sed 's/^/  /'
	@echo ""
	@echo "高级配置:"
	@echo "  BLDDIR         构建目录（默认: _output）"
	@echo "  VERSION        版本号（默认: 自动检测 Git 标签）"
	@echo "  BUILD_TIME     构建时间（默认: 当前时间）"
	@echo "  GIT_REVISION   Git 版本号（默认: 当前提交短哈希）"

# make VERSION="v1.2.3" BUILD_TIME="2023-10-01T00:00:00" build
# make BLDDIR=dist BUILD_PLATFORMS="linux/arm64" build

