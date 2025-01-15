package main

import (
	"embed"
	"os"
	"strings"

	"github.com/getlantern/elevate"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:front/dist
var assets embed.FS

func main() {

	args := os.Args
	// 使用 getlantern/elevate 来获取 root 权限
	needElevated := true
	for _, arg := range args {
		// wails bindings 时，不需要获取 root 权限
		if strings.HasSuffix(arg, "wailsbindings") {
			needElevated = false
			break
		}
		// 已取得 root 权限
		if arg == "--escalate" {
			needElevated = false
			break
		}
	}
	if needElevated {
		// 获取 root 权限
		elevate.Command(os.Args[0], "--escalate").Run()
		os.Exit(0)
	}

	// Create an instance of the app structure
	app := NewApp()
	// Create application with options
	err := wails.Run(&options.App{
		Title:  "platform-vpn",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		LogLevel:           logger.INFO,
		LogLevelProduction: logger.INFO,
		DisableResize:      true,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
