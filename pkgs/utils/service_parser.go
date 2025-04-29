package utils

import (
	"fmt"
	"strings"
)

// ServiceConfig 表示服务配置文件的结构
// 包含 Unit、Install 和 Service 三个主要配置段
type ServiceConfig struct {
	Unit    map[string]string // Unit 段配置
	Install map[string]string // Install 段配置
	Service map[string]string // Service 段配置
}

// ParseServiceFile 解析服务配置文件内容
//
// 参数:
//   - content: 服务配置文件的内容字符串
//
// 返回值:
//   - *ServiceConfig: 解析后的服务配置对象
//   - error: 解析过程中的错误，如果成功则返回 nil
func ParseServiceFile(content string) (*ServiceConfig, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("配置文件内容为空")
	}

	config := &ServiceConfig{
		Unit:    make(map[string]string),
		Install: make(map[string]string),
		Service: make(map[string]string),
	}

	var currentMap map[string]string
	var multiLineValue strings.Builder
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// 跳过注释行
		if strings.HasPrefix(line, "#") {
			continue
		}

		// 处理多行值
		if strings.HasSuffix(line, "\\") {
			multiLineValue.WriteString(strings.TrimSuffix(line, "\\"))
			multiLineValue.WriteString(" ")
			continue
		}

		// 如果存在未完成的多行值，添加当前行
		if multiLineValue.Len() > 0 {
			multiLineValue.WriteString(line)
			line = multiLineValue.String()
			multiLineValue.Reset()
		}

		// 处理配置段标记
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := line[1 : len(line)-1]
			switch section {
			case "Unit":
				currentMap = config.Unit
			case "Install":
				currentMap = config.Install
			case "Service":
				currentMap = config.Service
			default:
				return nil, fmt.Errorf("未知的配置段: %s", section)
			}
			continue
		}

		// 处理键值对
		if currentMap != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				if key == "" {
					return nil, fmt.Errorf("配置项键名不能为空")
				}
				
				currentMap[key] = value
			}
		}
	}

	// 验证必要的配置段是否存在
	if len(config.Unit) == 0 && len(config.Service) == 0 {
		return nil, fmt.Errorf("配置文件缺少必要的 [Unit] 或 [Service] 段")
	}

	return config, nil
}
