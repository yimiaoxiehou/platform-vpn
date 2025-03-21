wails build -platform darwin/arm64 -tags with_gvisor && cd build/bin && cp ../darwin/修复.sh . && zip -r platform-vpn-darwin.zip platform-vpn.app 修复.sh && rm -rf platform-vpn.app 修复.sh && cd ../..
wails build -platform windows/amd64 -tags with_gvisor -nsis
wails build -platform linux/amd64 -tags with_gvisor