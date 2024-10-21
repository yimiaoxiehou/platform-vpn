package util

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	Windows = "windows"
	Linux   = "linux"
	Darwin  = "darwin"
)

// FetchHosts 获取最新的host
func fetchHosts(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", err
	}

	fetchHosts, err := io.ReadAll(resp.Body)
	return string(fetchHosts), err
}

func getCleanPlatformHosts() (*bytes.Buffer, error) {
	hosts := &bytes.Buffer{}
	hostsPath := GetSystemHostsPath()
	hostsBytes, err := os.ReadFile(hostsPath)
	if err != nil {
		return hosts, err
	}
	hostsStr := string(hostsBytes)
	header := "## Platform START\n"
	end := "## Platform END\n"

	for {
		platformHostStartIdx := strings.Index(hostsStr, header)
		if platformHostStartIdx == -1 {
			hosts.WriteString(string(hostsBytes))
			break
		}

		platformHostEndIdx := strings.Index(hostsStr, end)
		if platformHostEndIdx == -1 {
			hosts.WriteString(string(hostsBytes))
			break
		}
		hostsStr = hostsStr[0:platformHostStartIdx] + hostsStr[platformHostEndIdx+len(end):]
	}

	_, err = hosts.Write([]byte(hostsStr))
	return hosts, err
}

// updatePlatformHosts 更新平台 Hosts 文件
func UpdatePlatformHosts(url string) error {
	hosts, err := getCleanPlatformHosts()
	if err != nil {
		return err
	}
	fetchHosts, err := fetchHosts(url)
	if err != nil {
		return err
	}
	_, err = hosts.WriteString(fetchHosts)
	if err != nil {
		return err
	}
	return os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType)
}

func CleanPlatformHosts() error {
	hosts, err := getCleanPlatformHosts()
	if err != nil {
		return err
	}
	return os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType)
}
