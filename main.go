package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/getlantern/elevate"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:gui/dist
var assets embed.FS

// 应用配置常量
const (
	AppTitle  = "platform-vpn"
	AppWidth  = 800
	AppHeight = 600
)

// checkElevation 检查并处理权限提升
func checkElevation() bool {
	args := os.Args
	for _, arg := range args {
		if strings.HasSuffix(arg, "wailsbindings") || arg == "--escalate" {
			return false
		}
	}
	return true
}

func main() {
	// 检查是否需要提升权限
	if checkElevation() {
		cmd := elevate.Command(os.Args[0], "--escalate")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "权限提升失败: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 创建应用实例
	app := NewApp()

	// 应用配置
	appConfig := &options.App{
		Title:  AppTitle,
		Width:  AppWidth,
		Height: AppHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
		BackgroundColour: &options.RGBA{
			R: 27,
			G: 38,
			B: 54,
			A: 1,
		},
		OnStartup:   app.startup,
		Bind:        []interface{}{app},
		LogLevel:    logger.INFO,
		DisableResize: false,
	}

	// 运行应用
	if err := wails.Run(appConfig); err != nil {
		fmt.Fprintf(os.Stderr, "应用启动失败: %v\n", err)
		os.Exit(1)
	}
}
