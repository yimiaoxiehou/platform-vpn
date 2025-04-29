package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// HostsBlockStart 标记平台 hosts 配置的开始
	HostsBlockStart = "## Platform START\n"
	// HostsBlockEnd 标记平台 hosts 配置的结束
	HostsBlockEnd = "## Platform END\n"
)

// GetSystemHostsPath 返回系统 hosts 文件的路径
//
// 根据当前操作系统返回对应的 hosts 文件路径：
// - Windows: %SystemRoot%\System32\drivers\etc\hosts
// - Unix/Linux/macOS: /etc/hosts
func GetSystemHostsPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

// getCleaningHosts 读取系统 hosts 文件并清除平台相关配置
//
// 返回值:
//   - *bytes.Buffer: 清理后的 hosts 内容
//   - error: 操作过程中的错误，如果成功则返回 nil
func getCleaningHosts() (*bytes.Buffer, error) {
	hosts := &bytes.Buffer{}
	hostsPath := GetSystemHostsPath()
	
	hostsBytes, err := os.ReadFile(hostsPath)
	if err != nil {
		return nil, fmt.Errorf("读取 hosts 文件失败: %w", err)
	}
	
	hostsStr := string(hostsBytes)
	platformHostStartIdx := strings.Index(hostsStr, HostsBlockStart)
	
	// 如果没有找到平台配置标记，直接返回原内容
	if platformHostStartIdx == -1 {
		if _, err := hosts.Write(hostsBytes); err != nil {
			return nil, fmt.Errorf("写入 hosts 内容失败: %w", err)
		}
		return hosts, nil
	}

	platformHostEndIdx := strings.Index(hostsStr, HostsBlockEnd)
	if platformHostEndIdx == -1 {
		if _, err := hosts.Write(hostsBytes); err != nil {
			return nil, fmt.Errorf("写入 hosts 内容失败: %w", err)
		}
		return hosts, nil
	}

	// 提取平台配置之前的内容
	hostsStr = strings.TrimSpace(hostsStr[:platformHostStartIdx])
	if _, err := hosts.WriteString(hostsStr); err != nil {
		return nil, fmt.Errorf("写入清理后的 hosts 内容失败: %w", err)
	}

	return hosts, nil
}

// UpdatePlatformHosts 更新平台 Hosts 文件
//
// 参数:
//   - appendHosts: 要追加的 hosts 配置内容
//
// 返回值:
//   - error: 更新过程中的错误，如果成功则返回 nil
func UpdatePlatformHosts(appendHosts string) error {
	hosts, err := getCleaningHosts()
	if err != nil {
		return fmt.Errorf("清理 hosts 文件失败: %w", err)
	}

	// 写入新的平台配置
	if _, err := hosts.WriteString("\n" + HostsBlockStart + appendHosts + HostsBlockEnd); err != nil {
		return fmt.Errorf("写入新 hosts 配置失败: %w", err)
	}

	// 使用正确的文件权限写入文件
	if err := os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), 0644); err != nil {
		return fmt.Errorf("保存 hosts 文件失败: %w", err)
	}

	return nil
}

// CleanPlatformHosts 清除平台相关的 Hosts 配置
//
// 返回值:
//   - error: 清理过程中的错误，如果成功则返回 nil
func CleanPlatformHosts() error {
	hosts, err := getCleaningHosts()
	if err != nil {
		return fmt.Errorf("清理 hosts 文件失败: %w", err)
	}

	// 使用正确的文件权限写入文件
	if err := os.WriteFile(GetSystemHostsPath(), hosts.Bytes(), 0644); err != nil {
		return fmt.Errorf("保存 hosts 文件失败: %w", err)
	}

	return nil
}
