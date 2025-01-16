#!/bin/bash

# 检查是否提供了新版本号
if [ -z "$1" ]; then
    echo "用法: $0 新版本号"
    exit 1
fi

# 方法: 判断系统类型
OS_TYPE=$(uname)
if [[ "$OS_TYPE" == "Darwin" ]]; then
    # macOS
    # 替换 wails.json 中的 productVersion
    sed -i "" "s|\"productVersion\": \".*\"|\"productVersion\": \"$1\"|" wails.json
    sed -i "" "s|\"version\": \".*\"|\"version\": \"$1\"|" gui/package.json
    sed -i "" "s|Platform VPN .*</h2>|Platform VPN $1</h2>|" gui/src/App.tsx
elif [[ "$OS_TYPE" == "Linux" ]]; then
    # Linux
    # 替换 wails.json 中的 productVersion
    sed -i "s|\"productVersion\": \".*\"|\"productVersion\": \"$1\"|" wails.json
    sed -i "s|\"version\": \".*\"|\"version\": \"$1\"|" gui/package.json
    sed -i "s|Platform VPN .*</h2>|Platform VPN $1</h2>|" gui/src/App.tsx
else
    echo "不支持的操作系统类型: $OS_TYPE"
    exit 1
fi

echo "版本号已更新为 $1" 

git add .
git commit -m "update version to $1"
git tag $1
git push -f origin $1
