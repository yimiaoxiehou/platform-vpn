wails build -platform darwin/arm64 -tags with_gvisor && cd build/bin && zip -r platform-vpn-darwin.zip platform-vpn.app && rm -rf platform-vpn.app && cd ../..
wails build -platform windows/amd64 -tags with_gvisor -nsis
wails build -platform linux/amd64 -tags with_gvisor