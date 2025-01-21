package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var HOST_START = "## Platform START\n"
var HOST_END = "## Platform END\n"

// GetHostsFilePath 返回系统 hosts 文件的路径
func GetSystemHostsPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

func getCleaningHosts() (*bytes.Buffer, error) {
	hosts := &bytes.Buffer{}
	hostsPath := GetSystemHostsPath()
	hostsBytes, err := os.ReadFile(hostsPath)
	if err != nil {
		return hosts, err
	}
	hostsStr := string(hostsBytes)

	platformHostStartIdx := strings.Index(hostsStr, HOST_START)
	if platformHostStartIdx == -1 {
		hosts.WriteString(string(hostsBytes))
		return hosts, err
	}

	platformHostEndIdx := strings.Index(hostsStr, HOST_END)
	if platformHostEndIdx == -1 {
		hosts.WriteString(string(hostsBytes))
		return hosts, err
	}
	hostsStr = hostsStr[0:platformHostStartIdx]
	hostsStr = strings.TrimSpace(hostsStr)

	_, err = hosts.Write([]byte(hostsStr))
	return hosts, err
}

// UpdatePlatformHosts 更新平台 Hosts 文件
func UpdatePlatformHosts(appendHosts string) error {
	hosts, err := getCleaningHosts()
	if err != nil {
		return err
	}
	_, err = hosts.WriteString("\n" + appendHosts)
	if err != nil {
		return err
	}
	return os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType)
}

func CleanPlatformHosts() error {
	hosts, err := getCleaningHosts()
	if err != nil {
		return err
	}
	return os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType)
}
