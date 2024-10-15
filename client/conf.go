package main

import (
	"github.com/spf13/viper"
)

type FetchConf struct {
	Lang         string
	Interval     int
	Method       string
	SelectOrigin string
	ServerAddr   string
	AutoFetch    bool
}

func (f *FetchConf) Storage() {
	viper.Set("lang", f.Lang)
	viper.Set("interval", f.Interval)
	viper.Set("serverAddr", f.ServerAddr)
	viper.Set("autofetch", f.AutoFetch)
	if err := viper.WriteConfigAs("conf.yaml"); err != nil {
		_fileLog.Print("持久化配置信息失败：" + err.Error())
	}
}

func LoadFetchConf() *FetchConf {
	viper.AddConfigPath(AppExecDir())
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.SetDefault("lang", "zh-CN")
	viper.SetDefault("interval", 60)
	viper.SetDefault("method", "官方指定hosts源")
	viper.SetDefault("selectorigin", "FetchGithubHosts")
	viper.SetDefault("autofetch", false)
	var fileNotExits bool
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fileNotExits = true
		} else {
			_fileLog.Print("加载配置文件错误： " + err.Error())
		}
	}
	res := FetchConf{}
	if err := viper.Unmarshal(&res); err != nil {
		_fileLog.Print("配置文件解析失败")
	}
	if fileNotExits {
		res.Storage()
	}
	return &res
}
