package main

import (
	"embed"
	"log"
	"os"
	"runtime"

	"github.com/getlantern/elevate"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

//go:embed assets
var assetsFs embed.FS

//go:embed active.*.toml
var localeFS embed.FS

var _cliLog = &FetchLog{w: os.Stdout}
var _local *i18n.Localizer
var _conf *FetchConf

func init() {
	// 设置日志文件
	logFile, err := os.OpenFile(AppExecDir()+"/fetch.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		_cliLog.Print(t(&i18n.Message{
			ID:    "LogCreatedFail",
			Other: "日志文件创建失败",
		}))
		return
	}
	defer logFile.Close()
	// 设置日志输出到多重写入器
	log.SetOutput(logFile)

	// 可选：设置日志前缀和标志
	log.SetPrefix("INFO: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	args := ParseBootArgs()
	if !args.DontEscalate && !args.Escalate && runtime.GOOS != Linux {
		cmd := elevate.Command(os.Args[0], "--escalate")
		cmd.Run()
		os.Exit(0)
	}
	bundle := i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFileFS(localeFS, "active.en-US.toml")
	_conf = LoadFetchConf()
	_local = i18n.NewLocalizer(bundle, args.Lang, _conf.Lang)

	bootGui()
}
