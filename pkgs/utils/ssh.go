package utils

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// SSHConfig SSH连接配置
type SSHConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// NewSSHClient 创建新的SSH客户端
func NewSSHClient(host string, port int, user string, password string) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %v", err)
	}

	return client, err
}
