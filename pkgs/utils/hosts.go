package utils

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var hostsMutex sync.Mutex

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
	header := "\n\n\n## Platform START\n\n\n"
	end := "\n\n\n## Platform END\n\n\n"

	platformHostStartIdx := strings.Index(hostsStr, header)
	if platformHostStartIdx == -1 {
		hosts.WriteString(string(hostsBytes))
		return hosts, err
	}

	platformHostEndIdx := strings.Index(hostsStr, end)
	if platformHostEndIdx == -1 {
		hosts.WriteString(string(hostsBytes))
		return hosts, err
	}
	hostsStr = hostsStr[0:platformHostStartIdx] + hostsStr[platformHostEndIdx+len(end):]

	_, err = hosts.Write([]byte(hostsStr))
	return hosts, err
}

// UpdatePlatformHosts 更新平台 Hosts 文件
func UpdatePlatformHosts(appendHosts string) error {
	hostsMutex.Lock()
	defer hostsMutex.Unlock()

	hosts, err := getCleaningHosts()
	if err != nil {
		return err
	}
	_, err = hosts.WriteString(appendHosts)
	if err != nil {
		return err
	}
	return os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType)
}

func CleanPlatformHosts(ctx context.Context) error {
	hostsMutex.Lock()
	defer hostsMutex.Unlock()

	hosts, err := getCleaningHosts()
	if err != nil {
		return err
	}
	return os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), os.ModeType)
}
