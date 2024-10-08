package main

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"strconv"
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
)

var mainWindow fyne.Window

var _fileLog *FetchLog

func bootGui() {
	logFile, err := os.OpenFile(AppExecDir()+"/fetch.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		_cliLog.Print(t(&i18n.Message{
			ID:    "LogCreatedFail",
			Other: "日志文件创建失败",
		}))
		return
	}
	_fileLog = &FetchLog{w: logFile}
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

	mainWindow.SetCloseIntercept(func() {
		mainWindow.Hide()
	})
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
	var interval, customUrl = strconv.Itoa(_conf.Interval), _conf.CustomUrl
	intervalInput, urlInput := widget.NewEntryWithData(binding.BindString(&interval)), widget.NewEntryWithData(binding.BindString(&customUrl))
	var ticker *FetchTicker

	intervalForm := widget.NewFormItem(t(&i18n.Message{
		ID:    "GetIntervalMinutes",
		Other: "获取间隔（分钟）",
	}), intervalInput)
	originForm := widget.NewFormItem(t(&i18n.Message{
		ID:    "RemoteHostsUrl",
		Other: "远程Hosts链接",
	}), urlInput)

	go startClient(ticker, customUrl, cLog)

	form := widget.NewForm(
		intervalForm,
		originForm,
	)

	startFetchExec := func() {
		intervalInt := parseStrIsNumberNotShowAlert(&interval, t(&i18n.Message{
			ID:    "GetIntervalNeedInt",
			Other: "获取间隔必须为整数",
		}))
		if intervalInt == nil {
			return
		}
		stopBtn.Enable()
		componentsStatusChange(false, startBtn, intervalInput, urlInput)
		ticker = NewFetchTicker(*intervalInt)
		go startClient(ticker, customUrl, cLog)

		_conf.CustomUrl = customUrl
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
		componentsStatusChange(true, startBtn, intervalInput, urlInput)
		ticker.Stop()
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

	buttons := container.New(layout.NewGridLayout(4), startBtn, stopBtn, widget.NewButton(t(&i18n.Message{
		ID:    "ClearHosts",
		Other: "清除hosts",
	}), func() {
		if err := flushCleanGithubHosts(); err != nil {
			showAlert(fmt.Sprintf("%s: %s", t(&i18n.Message{
				ID:    "CleanGithubHostsFail",
				Other: "清除hosts中的github记录失败",
			}), err.Error()))
		} else {
			showAlert(t(&i18n.Message{
				ID:    "CleanGithubHostsSuccess",
				Other: "hosts文件中的github记录已经清除成功",
			}))
		}
	}), container.New(layout.NewCenterLayout(), autoFetchCheck))
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

func openUrl(urlStr string) func() {
	return func() {
		u, _ := url.Parse(urlStr)
		_ = fyne.CurrentApp().OpenURL(u)
	}
}

func newMargin(size fyne.Size) *canvas.Rectangle {
	margin := canvas.NewRectangle(color.Transparent)
	margin.SetMinSize(size)
	return margin
}
