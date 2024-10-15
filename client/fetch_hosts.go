package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	Windows = "windows"
	Linux   = "linux"
	Darwin  = "darwin"
)

func startClient(ticker *FetchTicker, url string, flog *FetchLog) {
	flog.Print(tfs(&i18n.Message{
		ID:    "RemoteHostsUrlLog",
		Other: "远程hosts获取链接: {{.Url}}",
	}, map[string]interface{}{
		"Url": url,
	}))
	fn := func() {
		if err := ClientFetchHosts(url, flog); err != nil {
			flog.Print(tfs(&i18n.Message{
				ID:    "RemoteHostsFetchErrorLog",
				Other: "更新Platform-Hosts失败: {{.E}}",
			}, map[string]interface{}{
				"E": err.Error(),
			}))
		} else {
			flog.Print(t(&i18n.Message{
				ID:    "RemoteHostsFetchSuccessLog",
				Other: "更新Platform-Hosts成功！",
			}))
		}
	}
	fn()
	for {
		select {
		case <-ticker.Ticker.C:
			fn()
		case <-ticker.CloseChan:
			flog.Print(t(&i18n.Message{
				ID:    "RemoteHostsFetchStopLog",
				Other: "停止获取hosts",
			}))
			return
		}
	}
}

// ClientFetchHosts 获取最新的host并写入hosts文件
func ClientFetchHosts(url string, flog *FetchLog) (err error) {
	hosts, err := getCleanPlatformHosts()
	if err != nil {
		return
	}

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		err = ComposeError(t(&i18n.Message{
			ID:    "ClientFetchHostsGetErrorLog",
			Other: "获取最新的hosts失败",
		}), err)
		return
	}

	fetchHosts, err := io.ReadAll(resp.Body)
	if err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "ClientFetchHostsReadErrorLog",
			Other: "读取最新的hosts失败",
		}), err)
		return
	}
	hosts.Write(fetchHosts)
	if err = os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType); err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "WriteHostsNoPermission",
			Other: "写入hosts文件失败，请用超级管理员身份启动本程序！",
		}), err)
		return
	}

	return
}

func getCleanPlatformHosts() (hosts *bytes.Buffer, err error) {
	hosts = &bytes.Buffer{}
	hostsPath := GetSystemHostsPath()
	hostsBytes, err := os.ReadFile(hostsPath)
	if err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "ReadHostsErr",
			Other: "读取文件hosts错误",
		}), err)
		return
	}

	platformHostStartIdx := strings.Index(string(hostsBytes), "## Platform START\n")
	if platformHostStartIdx == -1 {
		hosts.WriteString(string(hostsBytes))
		return
	}

	platformHostEndIdx := strings.LastIndex(string(hostsBytes), "## Platform END\n")
	if platformHostEndIdx == -1 {
		hosts.WriteString(string(hostsBytes))
		return
	}

	hosts.Write(hostsBytes[0:platformHostStartIdx])
	hosts.Write(hostsBytes[platformHostEndIdx+16:])

	return
}

func flushCleanPlatformHosts() (err error) {
	hosts, err := getCleanPlatformHosts()
	if err != nil {
		return
	}
	if err = os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType); err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "WriteHostsNoPermission",
			Other: "写入hosts文件失败，请用超级管理员身份启动本程序！",
		}), err)
	}
	return
}
