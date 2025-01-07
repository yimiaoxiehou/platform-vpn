package utils

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
)

// GetSystemDNS 获取系统DNS服务器
func GetSystemDNS() ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsDNS()
	case "linux":
		return getLinuxDNS()
	case "darwin":
		return getDarwinDNS()
	default:
		return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// Windows系统获取DNS
func getWindowsDNS() ([]string, error) {
	cmd := exec.Command("ipconfig", "/all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行ipconfig命令失败: %v", err)
	}

	var dnsServers []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "DNS Servers") {
			server := strings.TrimSpace(strings.Split(line, ":")[1])
			if server != "" {
				dnsServers = append(dnsServers, server)
			}
		}
	}
	return dnsServers, nil
}

// Linux系统获取DNS
func getLinuxDNS() ([]string, error) {
	content, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("读取resolv.conf失败: %v", err)
	}

	var dnsServers []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		// 去掉注释和空行
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			server := strings.TrimSpace(strings.TrimPrefix(line, "nameserver"))
			if server != "" {
				dnsServers = append(dnsServers, server)
			}
		}
	}
	return dnsServers, nil
}

// macOS系统获取DNS
func getDarwinDNS() ([]string, error) {
	cmd := exec.Command("scutil", "--dns")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行scutil命令失败: %v", err)
	}

	var dnsServers []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "nameserver[") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				server := strings.TrimSpace(parts[1])
				if server != "" {
					dnsServers = append(dnsServers, server)
				}
			}
		}
	}
	return dnsServers, nil
}
