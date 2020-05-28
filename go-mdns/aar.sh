#!/bin/sh -e

# 设置目标文件夹
TARGET_DIR="/home/km/个人/android-iim/gomobile/"

echo "打包aar: ${PACKAGE}"

# 复制依赖
go mod vendor; cp -r vendor/* $GOPATH/src/; rm -rf $GOPATH/src/pkg $GOPATH/src/modules.txt ; rm -rf vendor

# 复制源码
TEMPPATH="$GOPATH/src/lilu.red/temp"
mkdir -p $TEMPPATH
# im包
cp -r im $TEMPPATH
# dns包
cp -r dns $TEMPPATH

# 打包
GO111MODULE="off"
gomobile bind -v -o "${TARGET_DIR}gomobile.aar" -target=android "${TEMPPATH}/im" "${TEMPPATH}/dns"
rm -rf TEMPPATH

echo "打包完成"