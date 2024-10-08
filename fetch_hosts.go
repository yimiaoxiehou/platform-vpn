package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

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
		if err := ClientFetchHosts(url); err != nil {
			flog.Print(tfs(&i18n.Message{
				ID:    "RemoteHostsFetchErrorLog",
				Other: "更新Github-Hosts失败: {{.E}}",
			}, map[string]interface{}{
				"E": err.Error(),
			}))
		} else {
			flog.Print(t(&i18n.Message{
				ID:    "RemoteHostsFetchSuccessLog",
				Other: "更新Github-Hosts成功！",
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
func ClientFetchHosts(url string) (err error) {
	hosts, err := getCleanGithubHosts()
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

	newlineChar := GetNewlineChar()

	fetchHostsLines := strings.Split(string(fetchHosts), "\n")

	for i, fetchLine := range fetchHostsLines {
		line := strings.TrimSpace(fetchLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		hosts.WriteString(fetchLine)
		if i != len(fetchHostsLines)-1 {
			hosts.WriteString(newlineChar)
		}
	}
	if err = os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType); err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "WriteHostsNoPermission",
			Other: "写入hosts文件失败，请用超级管理员身份启动本程序！",
		}), err)
		return
	}

	return
}

func FetchHosts(domains []string) (hostsJson, hostsFile []byte, now string, err error) {
	now = time.Now().Format("2006-01-02 15:04:05")
	hosts := make([][]string, 0, len(domains))
	hostsFileData := bytes.NewBufferString("# fetch-github-hosts begin\n")
	for _, domain := range domains {
		ipList, sErr := net.LookupIP(domain)
		if sErr != nil {
			fmt.Printf("%s: %s\b", t(&i18n.Message{
				ID:    "GetHostRecordErr",
				Other: "获取主机记录失败",
			}), sErr.Error())
			continue
		}
		for _, ip := range ipList {
			ipv4 := ip.To4()
			if ipv4 != nil {
				item := []string{ipv4.String(), domain}
				hosts = append(hosts, item)
				hostsFileData.WriteString(fmt.Sprintf("%-28s%s\n", item[0], item[1]))
				break
			}
		}
	}
	hostsFileData.WriteString("# last fetch time: ")
	hostsFileData.WriteString(now)
	hostsFileData.WriteString("\n# update url: https://hosts.gitcdn.top/hosts.txt\n# fetch-github-hosts end\n\n")
	hostsFile = hostsFileData.Bytes()
	hostsJson, err = json.Marshal(hosts)
	return
}

func getCleanGithubHosts() (hosts *bytes.Buffer, err error) {
	hostsPath := GetSystemHostsPath()
	hostsBytes, err := os.ReadFile(hostsPath)
	if err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "ReadHostsErr",
			Other: "读取文件hosts错误",
		}), err)
		return
	}

	domains, err := getGithubDomains()
	if err != nil {
		return
	}

	newlineChar := GetNewlineChar()

	// clear local hosts github domain
	localHostsLines := strings.Split(string(hostsBytes), newlineChar)
	hosts = &bytes.Buffer{}

	for i, localLine := range localHostsLines {
		line := strings.TrimSpace(localLine)
		if line == "" || strings.HasPrefix(line, "#") {
			hosts.WriteString(localLine)
			if i != len(localHostsLines)-1 || !strings.HasSuffix(hosts.String(), newlineChar) {
				hosts.WriteString(newlineChar)
			}
			continue
		}
		var clearLine bool
		for _, domain := range domains {
			if strings.Contains(line, domain) {
				clearLine = true
				break
			}
		}
		if !clearLine {
			hosts.WriteString(localLine)
			hosts.WriteString(newlineChar)
		}
	}

	return
}

func getGithubDomains() (domains []string, err error) {
	fileData, err := GetExecOrEmbedFile(&assetsFs, "assets/domains.json")
	if err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "ReadDomainsJsonErr",
			Other: "读取文件domains.json错误",
		}), err)
		return
	}

	if err = json.Unmarshal(fileData, &domains); err != nil {
		err = ComposeError(t(&i18n.Message{
			ID:    "ParseDomainsJsonErr",
			Other: "domain.json解析失败",
		}), err)
		return
	}
	return
}

func flushCleanGithubHosts() (err error) {
	hosts, err := getCleanGithubHosts()
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
