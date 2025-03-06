package utils

import (
	"strings"
)

type ServiceConfig struct {
	Unit    map[string]string
	Install map[string]string
	Service map[string]string
}

func ParseServiceFile(content string) (*ServiceConfig, error) {
	lines := strings.Split(content, "\n")
	lines = append(lines, "")

	config := &ServiceConfig{
		Unit:    make(map[string]string),
		Install: make(map[string]string),
		Service: make(map[string]string),
	}

	var currentMap map[string]string
	var multiLineValue strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasSuffix(line, "\\") {
			multiLineValue.WriteString(strings.TrimSuffix(line, "\\") + " ")
			continue
		} else {
			multiLineValue.WriteString(line)
			line = multiLineValue.String()
			multiLineValue.Reset()
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := line[1 : len(line)-1]
			switch section {
			case "Unit":
				currentMap = config.Unit
			case "Install":
				currentMap = config.Install
			case "Service":
				currentMap = config.Service
			}
			continue
		}

		if currentMap != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				currentMap[key] = value
			}
		}
	}
	return config, nil
}
