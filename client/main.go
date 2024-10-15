package main

import (
	"embed"
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
