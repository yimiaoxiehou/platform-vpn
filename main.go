package main

import (
	"embed"
	"log"
	"os"
	"runtime"

	"github.com/Licoy/fetch-github-hosts/util"
	"github.com/getlantern/elevate"
)

//go:embed assets
var assetsFs embed.FS

var _cliLog = util.NewFetchLog(os.Stdout)
var _conf *util.FetchConf

func init() {
	// 设置日志文件
	logFile, err := os.OpenFile(util.AppExecDir()+"/fetch.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		_cliLog.Print("日志文件创建失败")
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
	args := util.ParseBootArgs(VERSION)
	if !args.DontEscalate && !args.Escalate && runtime.GOOS != util.Linux {
		cmd := elevate.Command(os.Args[0], "--escalate")
		cmd.Run()
		os.Exit(0)
	}
	_conf = util.LoadFetchConf()
	bootGui()
}
