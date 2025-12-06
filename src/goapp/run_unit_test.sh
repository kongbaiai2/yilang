#!/bin/bash

APP="goapp"

ROOTDIR=`pwd`
COVERDIR="$ROOTDIR/cover/"

function pre_init() {
    mkdir -p $COVERDIR
    #chmod +x $ROOTDIR/$APP/tools/gocov
    #chmod +x $ROOTDIR/$APP/tools/gocov-xml

    #sudo pip install diff-cover==2.6.1
}

function run_unit_test() {
    # 添加你的单元测试命令，使用gomonkey需要关闭内联编译
    (go test -v -gcflags=-l -cover -coverprofile $COVERDIR/main.cover ./...)

    #/api_server/cover/pkgA.cover
    #/api_server/cover/*.cover

    # 按pkg mock进行规范化打包
    #(cd $APP/mock && go test -v -gcflags=-l -cover -coverprofile $COVERDIR/mock.cover -run .)
    #(cd $APP/mock && go test -v -gcflags=-l -cover -coverprofile $COVERDIR/mock.cover -run .)
    #(cd $APP/mock && go test -v -gcflags=-l -cover -coverprofile $COVERDIR/mock.cover -run .)
}

function run_diff_cover() {
    # 全局main包增量覆盖率
    $tools/gocov convert $COVERDIR/main.cover | tools/gocov-xml > $COVERDIR/main.xml

    # pkg包增量覆盖率
    #$ROOTDIR/$APP/tools/mock convert $COVERDIR/mock.cover | $ROOTDIR/$APP/tools/gocov-xml > $COVERDIR/mock.xml

    # 增量覆盖率 fail-under如果小于10则失败
    diff-cover main.xml --compare-branch=origin/master --html-report cover/diff-report.html --fail-under=1 
}

pre_init
run_unit_test
#run_diff_cover

