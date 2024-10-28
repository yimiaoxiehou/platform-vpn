package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/yimiaoxiehou/platform-vpn/util"
	"github.com/yimiaoxiehou/tun2socks/core"
)

var mainWindow fyne.Window
var engine core.Engine
var hostsUrl string

func bootGui() {
	logoResource := getLogoResource()
	a := app.New()
	a.Settings().SetTheme(&fghGuiTheme{})

	mainWindow = a.NewWindow(fmt.Sprintf("UT Platform host fetch - V%.1f", VERSION))
	mainWindow.Resize(fyne.NewSize(600, 400))
	mainWindow.SetIcon(logoResource)

	logoImage := canvas.NewImageFromResource(logoResource)
	logoImage.SetMinSize(fyne.NewSize(240, 240))

	if err := util.GetCheckPermissionResult(); err != nil {
		time.AfterFunc(time.Second, func() {
			showAlert(err.Error())
		})
	}
	mainWindow.CenterOnScreen()
	mainWindow.SetContent(guiClientMode())
	mainWindow.ShowAndRun()
}

func getLogoResource() fyne.Resource {
	content, err := assetsFs.ReadFile("assets/public/logo.png")
	if err != nil {
		return nil
	}
	return &fyne.StaticResource{StaticName: "logo", StaticContent: content}
}

func guiClientMode() (content fyne.CanvasObject) {
	logs, addFn := newLogScrollComponent(fyne.NewSize(590, 280))
	var cLog = util.NewFetchLog(util.NewGuiLogWriter(addFn))
	var startBtn, stopBtn, refreshBtn *widget.Button
	var interval = strconv.Itoa(_conf.Interval)
	serverAddr := _conf.ServerAddr
	intervalInput := widget.NewEntryWithData(binding.BindString(&interval))
	serverInput := widget.NewEntryWithData(binding.BindString(&serverAddr))
	var ticker *util.FetchTicker

	intervalForm := widget.NewFormItem("获取间隔（分钟）", intervalInput)
	serverForm := widget.NewFormItem("服务器地址", serverInput)

	form := widget.NewForm(
		intervalForm,
		serverForm,
	)

	startExec := func() {
		if serverAddr == "" {
			return
		}
		if len(strings.Split(serverAddr, ":")) == 1 {
			serverAddr += ":1080"
		}

		hostsUrl = "http://" + serverAddr + "/hosts"
		intervalInt := parseStrIsNumberNotShowAlert(&interval, "获取间隔必须为整数")
		if intervalInt == nil {
			return
		}
		stopBtn.Enable()
		componentsStatusChange(false, startBtn, intervalInput, serverInput)
		ticker = util.NewFetchTicker(*intervalInt)
		nets, err := util.GetPlatformNets("http://" + serverAddr + "/nets")
		if err != nil {
			cLog.Print("获取平台网络失败: " + err.Error())
			return
		}
		go func() {
			engine = core.Engine{
				TunDevice: "utpf-tun",
				TunAddr:   "10.10.10.10",
				TunMask:   "255.255.255.255",
				Mtu:       1420,
				Sock5Addr: "socks5://" + serverAddr,
				Routers:   nets,
			}
			err := engine.Start()
			if err != nil {
				cLog.Print(err.Error())
			}
		}()
		go func() {
			cLog.Print("远程hosts获取链接: " + hostsUrl)

			fn := func() {
				err := util.UpdatePlatformHosts(hostsUrl)
				if err != nil {
					cLog.Print("更新Platform-Hosts失败: " + err.Error())
				} else {
					cLog.Print("更新Platform-Hosts成功！")
				}
			}
			fn()
			for {
				select {
				case <-ticker.Ticker.C:
					fn()
				case <-ticker.CloseChan:
					cLog.Print("停止获取hosts")
					return
				}
			}
		}()

		_conf.ServerAddr = serverAddr
		_conf.Interval = *intervalInt
		_conf.Storage()
	}
	stopExec := func() {
		stopBtn.Disable()
		componentsStatusChange(true, startBtn, intervalInput, serverInput)
		engine.Stop()
		ticker.Stop()
		err := util.CleanPlatformHosts()
		if err != nil {
			cLog.Print("清理Platform-Hosts失败: " + err.Error())
		} else {
			cLog.Print("清理Platform-Hosts成功！")
		}
	}
	refreshExec := func() {
		if startBtn.Disabled() {
			err := util.UpdatePlatformHosts(hostsUrl)
			if err != nil {
				cLog.Print("更新Platform-Hosts失败: " + err.Error())
			} else {
				cLog.Print("更新Platform-Hosts成功！")
			}
		}
	}

	startBtn = widget.NewButton("启动", startExec)
	stopBtn = widget.NewButton("停止", stopExec)
	refreshBtn = widget.NewButton("刷新 Hosts", refreshExec)

	if _conf.AutoFetch {
		startExec()
		startBtn.Disable()
	} else {
		stopBtn.Disable()
	}
	autoFetchCheck := widget.NewCheck("启动软件自动获取", func(b bool) {
		if b != _conf.AutoFetch {
			_conf.AutoFetch = b
			_conf.Storage()
			showAlert("启动软件自动获取状态已改变，将会在下次启动程序时生效！")
		}
	})
	autoFetchCheck.SetChecked(_conf.AutoFetch)

	buttons := container.New(layout.NewGridLayout(4), startBtn, stopBtn, refreshBtn, container.New(layout.NewCenterLayout(), autoFetchCheck))
	margin := newMargin(fyne.NewSize(10, 10))
	return container.NewVBox(margin, form, margin, buttons, margin, logs)
}

func showAlert(msg string) {
	dialog.NewCustom("提示", "确认", widget.NewLabel(msg), mainWindow).Show()
}

func parseStrIsNumberNotShowAlert(str *string, msg string) *int {
	res, err := strconv.Atoi(*str)
	if err != nil {
		showAlert(msg)
		return nil
	}
	return &res
}

func newLogScrollComponent(size fyne.Size) (scroll *container.Scroll, addFn func(string)) {
	var logs string
	textarea := widget.NewMultiLineEntry()
	textarea.Wrapping = fyne.TextWrapBreak
	textarea.Disable()
	scroll = container.NewScroll(textarea)
	scroll.SetMinSize(size)
	addFn = func(s string) {
		logs = s + logs
		textarea.SetText(logs)
	}
	return
}

func componentsStatusChange(enable bool, components ...fyne.Disableable) {
	for _, v := range components {
		if enable {
			v.Enable()
		} else {
			v.Disable()
		}
	}
}

func newMargin(size fyne.Size) *canvas.Rectangle {
	margin := canvas.NewRectangle(color.Transparent)
	margin.SetMinSize(size)
	return margin
}
