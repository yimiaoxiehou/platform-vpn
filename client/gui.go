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
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/yimiaoxiehou/tun2socks/core"
)

var mainWindow fyne.Window
var _fileLog *FetchLog
var engine core.Engine

func bootGui() {
	logoResource := getLogoResource()
	a := app.New()
	a.Settings().SetTheme(&fghGuiTheme{})

	mainWindow = a.NewWindow(fmt.Sprintf("UT Platform host fetch - V%.1f", VERSION))
	mainWindow.Resize(fyne.NewSize(800, 580))
	mainWindow.SetIcon(logoResource)

	logoImage := canvas.NewImageFromResource(logoResource)
	logoImage.SetMinSize(fyne.NewSize(240, 240))

	if err := GetCheckPermissionResult(); err != nil {
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

func getTicker(interval int) *time.Ticker {
	d := time.Minute
	if IsDebug() {
		d = time.Second
	}
	return time.NewTicker(d * time.Duration(interval))
}

func guiClientMode() (content fyne.CanvasObject) {
	logs, addFn := newLogScrollComponent(fyne.NewSize(800, 280))
	var cLog = NewFetchLog(NewGuiLogWriter(addFn))
	var startBtn, stopBtn *widget.Button
	var interval = strconv.Itoa(_conf.Interval)
	serverAddr := _conf.ServerAddr
	intervalInput := widget.NewEntryWithData(binding.BindString(&interval))
	serverInput := widget.NewEntryWithData(binding.BindString(&serverAddr))
	var ticker *FetchTicker

	intervalForm := widget.NewFormItem(t(&i18n.Message{
		ID:    "GetIntervalMinutes",
		Other: "获取间隔（分钟）",
	}), intervalInput)
	serverForm := widget.NewFormItem(t(&i18n.Message{
		ID:    "ServerAddr",
		Other: "服务器地址",
	}), serverInput)

	form := widget.NewForm(
		intervalForm,
		serverForm,
	)

	startFetchExec := func() {
		if serverAddr == "" {
			return
		}
		if len(strings.Split(serverAddr, ":")) == 1 {
			serverAddr += ":1080"
		}
		intervalInt := parseStrIsNumberNotShowAlert(&interval, t(&i18n.Message{
			ID:    "GetIntervalNeedInt",
			Other: "获取间隔必须为整数",
		}))
		if intervalInt == nil {
			return
		}
		stopBtn.Enable()
		componentsStatusChange(false, startBtn, intervalInput, serverInput)
		ticker = NewFetchTicker(*intervalInt)
		go func() {
			engine = core.Engine{
				TunDevice: "utpf-tun",
				TunAddr:   "10.96.255.255",
				TunMask:   "255.240.0.0",
				Mtu:       1420,
				Sock5Addr: "socks5://" + serverAddr,
			}
			err := engine.Start()
			if err != nil {
				cLog.Print(err.Error())
			}
		}()
		go startClient(ticker, "http://"+serverAddr, cLog)

		_conf.ServerAddr = serverAddr
		_conf.Interval = *intervalInt
		_conf.Storage()
	}

	startBtn = widget.NewButton(t(&i18n.Message{
		ID:    "Start",
		Other: "启动",
	}), startFetchExec)
	stopBtn = widget.NewButton(t(&i18n.Message{
		ID:    "Stop",
		Other: "停止",
	}), func() {
		stopBtn.Disable()
		componentsStatusChange(true, startBtn, intervalInput, serverInput)
		engine.Stop()
		ticker.Stop()
		if err := flushCleanPlatformHosts(); err != nil {
			cLog.Print(fmt.Sprintf("%s: %s", t(&i18n.Message{
				ID:    "CleanHostsFail",
				Other: "清除hosts中的 platform 记录失败",
			}), err.Error()))
		} else {
			cLog.Print(t(&i18n.Message{
				ID:    "CleanHostsSuccess",
				Other: "hosts文件中的 platform 记录已经清除成功",
			}))
		}
	})

	if _conf.AutoFetch {
		startFetchExec()
		startBtn.Disable()
	} else {
		stopBtn.Disable()
	}
	autoFetchCheck := widget.NewCheck(t(&i18n.Message{
		ID:    "StartupAutoGet",
		Other: "启动软件自动获取",
	}), func(b bool) {
		if b != _conf.AutoFetch {
			_conf.AutoFetch = b
			_conf.Storage()
			showAlert(t(&i18n.Message{
				ID:    "StartupAutoGetTips",
				Other: "启动软件自动获取状态已改变，将会在下次启动程序时生效！",
			}))
		}
	})
	autoFetchCheck.SetChecked(_conf.AutoFetch)

	buttons := container.New(layout.NewGridLayout(4), startBtn, stopBtn, container.New(layout.NewCenterLayout(), autoFetchCheck))
	margin := newMargin(fyne.NewSize(10, 10))
	return container.NewVBox(margin, form, margin, buttons, margin, logs)
}

func showAlert(msg string) {
	dialog.NewCustom(t(&i18n.Message{
		ID:    "Tip",
		Other: "提示",
	}), t(&i18n.Message{
		ID:    "Ok",
		Other: "确认",
	}), widget.NewLabel(msg), mainWindow).Show()
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
