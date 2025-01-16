#!/bin/bash

# 检查是否提供了新版本号
if [ -z "$1" ]; then
    echo "用法: $0 新版本号"
    exit 1
fi

# 替换 wails.json 中的 productVersion
sed -i "s/\"productVersion\": \".*\"/\"productVersion\": \"$1\"/" wails.json
sed -i "s/\"version\": \".*\"/\"version\": \"$1\"/" gui/package.json
echo "版本号已更新为 $1" 

git add .
git commit -m "update version to $1"
git tag $1
git push origin $1
