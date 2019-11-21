#!/bin/sh -e

PROJECT="github.com/alx696/go-dht-fire-mp2p"
PROJECT_PATH="$GOPATH/src/$PROJECT"

echo "打包mp2p aar"

# 复制源码
mkdir -p $PROJECT_PATH
cp -r mp2p $PROJECT_PATH

# 复制依赖
go mod vendor; cp -r vendor/* $GOPATH/src/; rm -rf $GOPATH/src/pkg $GOPATH/src/modules.txt ; rm -rf vendor

# 打包(仅arm64)
GO111MODULE="off"
gomobile bind -v -o mp2p.aar -target=android/arm64 $PROJECT/mp2p
rm -rf $PROJECT

echo "打包完成"